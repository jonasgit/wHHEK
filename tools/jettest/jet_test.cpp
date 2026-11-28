/*
 * JETTEST.EXE - Win32 test that MDAC and the Jet 4.0 database engine work in
 * a Wine prefix (the goal of the winetricks patch that lets mdac27/mdac28/
 * jet40 install on new-style WoW64 prefixes).
 *
 * Jet 4.0 is 32-bit only, so this is a 32-bit program; on a 64-bit prefix it
 * runs under WoW64.  It checks:
 *
 *  1. MDAC is registered: HKLM\Software\Microsoft\DataAccess (32-bit view).
 *  2. ADO works: ADODB.Connection can be created, and it comes from MDAC's
 *     msado15.dll, not Wine's builtin partial implementation.
 *  3. Jet 4.0 works end to end: create a new .mdb with ADOX.Catalog through
 *     the Microsoft.Jet.OLEDB.4.0 provider, create a table, insert a row and
 *     read it back.
 *  4. The ODBC Access driver works: connect to that database with
 *     "Driver={Microsoft Access Driver (*.mdb)};DBQ=<file>" (as programs using
 *     ODBC do, e.g. wHHEK) and read the row back.
 *
 * Without the patch, winetricks refuses jet40 on new-style WoW64 (or MDAC's
 * setup installs nothing), and tests 1 and 3 fail.
 *
 * The report goes to stdout and to a message box (any command line argument,
 * e.g. /q, suppresses the box).  The exit code is the number of failed tests.
 *
 * Build with MinGW: see build.sh.
 */

#include <windows.h>
#include <ole2.h>
#include <sql.h>
#include <sqlext.h>
#include <stdarg.h>
#include <stdio.h>
#include <wchar.h>

static char report[8192];
static int failures, test_failed;

static void out( const char *fmt, ... )
{
    size_t len = strlen( report );
    va_list args;

    va_start( args, fmt );
    vsnprintf( report + len, sizeof(report) - len, fmt, args );
    va_end( args );
    fputs( report + len, stdout );
    fflush( stdout );
}

static void begin_test( int n, const char *name )
{
    test_failed = 0;
    out( "%d. %s\n", n, name );
}

static void end_test( void )
{
    if (test_failed) failures++;
    out( "   => %s\n\n", test_failed ? "FAIL" : "PASS" );
}

static bool check( bool cond, const char *fmt, ... )
{
    char buf[512];
    va_list args;

    va_start( args, fmt );
    vsnprintf( buf, sizeof(buf), fmt, args );
    va_end( args );
    if (!cond) test_failed = 1;
    out( "   %-4s %s\n", cond ? "ok" : "FAIL", buf );
    return cond;
}

static const char *narrow( const WCHAR *s )
{
    static char bufs[4][512];
    static int n;
    char *buf = bufs[n++ % 4];

    if (!s) return "(null)";
    WideCharToMultiByte( CP_UTF8, 0, s, -1, buf, sizeof(bufs[0]), NULL, NULL );
    return buf;
}

/* the COM error description, if the failing object set one */
static const char *error_text( HRESULT hr )
{
    static char buf[600];
    IErrorInfo *info = NULL;
    BSTR desc = NULL;

    snprintf( buf, sizeof(buf), "hr %08lx", hr );
    if (GetErrorInfo( 0, &info ) == S_OK && info)
    {
        if (SUCCEEDED( info->GetDescription( &desc ) ) && desc && *desc)
            snprintf( buf, sizeof(buf), "hr %08lx: %s", hr, narrow( desc ) );
        SysFreeString( desc );
        info->Release();
    }
    return buf;
}

/* IDispatch::Invoke by name; args are given in natural (left to right) order */
static HRESULT invoke( IDispatch *disp, const WCHAR *name, WORD flags, VARIANT *result, int argc = 0,
                       VARIANT *args = NULL )
{
    DISPID id, put_id = DISPID_PROPERTYPUT;
    DISPPARAMS params = { 0 };
    VARIANT rev[4];
    LPOLESTR names[1] = { (LPOLESTR)name };
    HRESULT hr;
    int i;

    if (result) VariantInit( result );
    hr = disp->GetIDsOfNames( IID_NULL, names, 1, LOCALE_USER_DEFAULT, &id );
    if (FAILED( hr )) return hr;
    for (i = 0; i < argc; i++) rev[i] = args[argc - 1 - i];
    params.rgvarg = rev;
    params.cArgs = argc;
    if (flags & DISPATCH_PROPERTYPUT)
    {
        params.rgdispidNamedArgs = &put_id;
        params.cNamedArgs = 1;
    }
    return disp->Invoke( id, IID_NULL, LOCALE_USER_DEFAULT, flags, &params, result, NULL, NULL );
}

static VARIANT bstr_arg( const WCHAR *s )
{
    VARIANT v;
    V_VT(&v) = VT_BSTR;
    V_BSTR(&v) = SysAllocString( s );
    return v;
}

static IDispatch *create_object( const WCHAR *progid, HRESULT *ret )
{
    IDispatch *disp = NULL;
    CLSID clsid;

    *ret = CLSIDFromProgID( progid, &clsid );
    if (SUCCEEDED( *ret ))
        *ret = CoCreateInstance( clsid, NULL, CLSCTX_INPROC_SERVER, IID_IDispatch, (void **)&disp );
    return disp;
}

static void test_mdac_registry( void )
{
    WCHAR ver[64] = L"";
    DWORD size = sizeof(ver);
    LONG ret;

    begin_test( 1, "MDAC is registered" );
    ret = RegGetValueW( HKEY_LOCAL_MACHINE, L"Software\\Microsoft\\DataAccess", L"FullInstallVer",
                        RRF_RT_REG_SZ, NULL, ver, &size );
    check( ret == ERROR_SUCCESS && *ver, "HKLM\\Software\\Microsoft\\DataAccess FullInstallVer: %s",
           ret == ERROR_SUCCESS ? narrow( ver ) : "(missing)" );
    end_test();
}

static void test_ado( void )
{
    IDispatch *conn;
    VARIANT v;
    HRESULT hr;

    begin_test( 2, "ADO (ADODB.Connection)" );
    conn = create_object( L"ADODB.Connection", &hr );
    if (check( conn != NULL, "create ADODB.Connection: %s", error_text( hr ) ))
    {
        HMODULE mod = GetModuleHandleW( L"msado15.dll" );
        WCHAR file[MAX_PATH] = L"";

        hr = invoke( conn, L"Version", DISPATCH_PROPERTYGET, &v );
        check( hr == S_OK && V_VT(&v) == VT_BSTR, "Connection.Version: %s",
               hr == S_OK && V_VT(&v) == VT_BSTR ? narrow( V_BSTR(&v) ) : error_text( hr ) );
        VariantClear( &v );

        /* Wine's own msado15 is a partial stand-in; MDAC installs Microsoft's.  Wine marks
         * its builtin modules with "Wine builtin DLL" where the DOS stub normally is. */
        if (mod) GetModuleFileNameW( mod, file, MAX_PATH );
        check( mod && memcmp( (char *)mod + 0x40, "Wine builtin DLL", 16 ),
               "msado15.dll is MDAC's, not Wine's builtin: %s%s", narrow( file ),
               mod && !memcmp( (char *)mod + 0x40, "Wine builtin DLL", 16 ) ? " (Wine builtin)" : "" );
        conn->Release();
    }
    end_test();
}

static WCHAR db_path[MAX_PATH];  /* database created by test 3, read by test 4 */

static void test_jet( void )
{
    WCHAR *path = db_path, connstr[MAX_PATH + 64];
    IDispatch *catalog, *conn = NULL, *rs, *fields, *field;
    VARIANT v, args[2], result;
    HRESULT hr;

    begin_test( 3, "Jet 4.0 database round trip (Microsoft.Jet.OLEDB.4.0)" );
    GetTempPathW( MAX_PATH - 16, path );
    wcscat( path, L"jettest.mdb" );
    DeleteFileW( path );
    swprintf( connstr, sizeof(connstr) / sizeof(WCHAR), L"Provider=Microsoft.Jet.OLEDB.4.0;Data Source=%ls", path );

    /* create a new database file */
    catalog = create_object( L"ADOX.Catalog", &hr );
    if (!check( catalog != NULL, "create ADOX.Catalog: %s", error_text( hr ) )) { end_test(); return; }
    args[0] = bstr_arg( connstr );
    hr = invoke( catalog, L"Create", DISPATCH_METHOD, &result, 1, args );
    VariantClear( &args[0] );
    VariantClear( &result );
    catalog->Release();
    if (!check( hr == S_OK && GetFileAttributesW( path ) != INVALID_FILE_ATTRIBUTES,
                "Catalog.Create %s: %s", narrow( path ), error_text( hr ) )) { end_test(); return; }

    conn = create_object( L"ADODB.Connection", &hr );
    if (!check( conn != NULL, "create ADODB.Connection: %s", error_text( hr ) )) { end_test(); return; }
    args[0] = bstr_arg( connstr );
    hr = invoke( conn, L"Open", DISPATCH_METHOD, NULL, 1, args );
    VariantClear( &args[0] );
    if (!check( hr == S_OK, "Connection.Open: %s", error_text( hr ) )) { conn->Release(); end_test(); return; }

    args[0] = bstr_arg( L"CREATE TABLE greeting (id INTEGER, txt TEXT(50))" );
    hr = invoke( conn, L"Execute", DISPATCH_METHOD, &result, 1, args );
    VariantClear( &args[0] );
    VariantClear( &result );
    check( hr == S_OK, "CREATE TABLE: %s", error_text( hr ) );

    args[0] = bstr_arg( L"INSERT INTO greeting VALUES (1, 'hello from Jet')" );
    hr = invoke( conn, L"Execute", DISPATCH_METHOD, &result, 1, args );
    VariantClear( &args[0] );
    VariantClear( &result );
    check( hr == S_OK, "INSERT: %s", error_text( hr ) );

    args[0] = bstr_arg( L"SELECT txt FROM greeting WHERE id = 1" );
    hr = invoke( conn, L"Execute", DISPATCH_METHOD, &result, 1, args );
    VariantClear( &args[0] );
    if (check( hr == S_OK && V_VT(&result) == VT_DISPATCH, "SELECT: %s", error_text( hr ) ))
    {
        rs = V_DISPATCH(&result);
        hr = invoke( rs, L"Fields", DISPATCH_PROPERTYGET, &v );
        fields = (hr == S_OK && V_VT(&v) == VT_DISPATCH) ? V_DISPATCH(&v) : NULL;
        if (fields)
        {
            V_VT(&args[0]) = VT_I4;
            V_I4(&args[0]) = 0;
            hr = invoke( fields, L"Item", DISPATCH_PROPERTYGET | DISPATCH_METHOD, &v, 1, args );
            field = (hr == S_OK && V_VT(&v) == VT_DISPATCH) ? V_DISPATCH(&v) : NULL;
            if (field)
            {
                hr = invoke( field, L"Value", DISPATCH_PROPERTYGET, &v );
                check( hr == S_OK && V_VT(&v) == VT_BSTR && !wcscmp( V_BSTR(&v), L"hello from Jet" ),
                       "read back: \"%s\" (want \"hello from Jet\")",
                       hr == S_OK && V_VT(&v) == VT_BSTR ? narrow( V_BSTR(&v) ) : error_text( hr ) );
                VariantClear( &v );
                field->Release();
            }
            else check( false, "Fields.Item(0): %s", error_text( hr ) );
            fields->Release();
        }
        else check( false, "Recordset.Fields: %s", error_text( hr ) );
    }
    VariantClear( &result );
    invoke( conn, L"Close", DISPATCH_METHOD, NULL );
    conn->Release();
    end_test();
}

/* the ODBC diagnostic message of the last failed call on a handle */
static const char *odbc_error( SQLSMALLINT type, SQLHANDLE handle, SQLRETURN ret )
{
    static char buf[600];
    SQLWCHAR state[6] = { 0 }, msg[400] = { 0 };
    SQLINTEGER native = 0;
    SQLSMALLINT len;

    snprintf( buf, sizeof(buf), "ret %d", ret );
    if (handle && SQL_SUCCEEDED( SQLGetDiagRecW( type, handle, 1, state, &native, msg, 400, &len ) ))
        snprintf( buf, sizeof(buf), "ret %d: [%s] %s", ret, narrow( (WCHAR *)state ), narrow( (WCHAR *)msg ) );
    return buf;
}

static void test_odbc( void )
{
    SQLHENV env = SQL_NULL_HENV;
    SQLHDBC dbc = SQL_NULL_HDBC;
    SQLHSTMT stmt = SQL_NULL_HSTMT;
    WCHAR connstr[MAX_PATH + 64], txt[64] = L"";
    SQLLEN len = 0;
    SQLRETURN ret;

    begin_test( 4, "ODBC: Microsoft Access Driver (*.mdb)" );
    if (!check( GetFileAttributesW( db_path ) != INVALID_FILE_ATTRIBUTES, "database from test 3 exists" ))
    {
        end_test();
        return;
    }
    swprintf( connstr, sizeof(connstr) / sizeof(WCHAR), L"Driver={Microsoft Access Driver (*.mdb)};DBQ=%ls", db_path );

    SQLAllocHandle( SQL_HANDLE_ENV, SQL_NULL_HANDLE, &env );
    SQLSetEnvAttr( env, SQL_ATTR_ODBC_VERSION, (SQLPOINTER)SQL_OV_ODBC3, 0 );
    SQLAllocHandle( SQL_HANDLE_DBC, env, &dbc );
    ret = SQLDriverConnectW( dbc, NULL, (SQLWCHAR *)connstr, SQL_NTS, NULL, 0, NULL, SQL_DRIVER_NOPROMPT );
    if (check( SQL_SUCCEEDED( ret ), "SQLDriverConnect: %s", SQL_SUCCEEDED( ret ) ? "ok" : odbc_error( SQL_HANDLE_DBC, dbc, ret ) ))
    {
        SQLAllocHandle( SQL_HANDLE_STMT, dbc, &stmt );
        ret = SQLExecDirectW( stmt, (SQLWCHAR *)L"SELECT txt FROM greeting WHERE id = 1", SQL_NTS );
        if (check( SQL_SUCCEEDED( ret ), "SELECT: %s", SQL_SUCCEEDED( ret ) ? "ok" : odbc_error( SQL_HANDLE_STMT, stmt, ret ) ))
        {
            ret = SQLFetch( stmt );
            if (SQL_SUCCEEDED( ret ))
                ret = SQLGetData( stmt, 1, SQL_C_WCHAR, txt, sizeof(txt), &len );
            check( SQL_SUCCEEDED( ret ) && !wcscmp( txt, L"hello from Jet" ), "read back: \"%s\" (want \"hello from Jet\")",
                   SQL_SUCCEEDED( ret ) ? narrow( txt ) : odbc_error( SQL_HANDLE_STMT, stmt, ret ) );
        }
        SQLFreeHandle( SQL_HANDLE_STMT, stmt );
        SQLDisconnect( dbc );
    }
    SQLFreeHandle( SQL_HANDLE_DBC, dbc );
    SQLFreeHandle( SQL_HANDLE_ENV, env );
    end_test();
}

int main( int argc, char **argv )
{
    BOOL wow64 = FALSE;

    CoInitialize( NULL );
    IsWow64Process( GetCurrentProcess(), &wow64 );
    out( "MDAC / Jet 4.0 in this prefix\n\n32-bit process%s\n\n", wow64 ? " under WoW64 (64-bit prefix)" : "" );

    test_mdac_registry();
    test_ado();
    test_jet();
    test_odbc();
    if (*db_path) DeleteFileW( db_path );

    out( "Result: %s\n", failures ? "FAIL" : "PASS" );
    if (argc < 2)
        MessageBoxA( NULL, report, "MDAC / Jet test", MB_OK | (failures ? MB_ICONSTOP : MB_ICONINFORMATION) );
    CoUninitialize();
    return failures;
}
