/*
-----------------------------------------------------------

    @Filename:         main.go
    @Copyright Author: Yogesh K
    @Date:             21/02/2023

-------------------------------------------------------------
*/
package main

import (
    "database/sql"
    "log"
    "net/http"
    "flag"
    "os"
    // Import the models package from internal/models.
    "github.com/cpucortexm/chunkbox/internal/models"
    _ "github.com/go-sql-driver/mysql" //we need the driver’s init() function to run so that it can register itself with the database/sql package.
)

// Define an application struct to hold the application-wide dependencies for the
// web application. For now we'll only include fields for the two custom loggers.
type application struct {
    errorLog *log.Logger
    infoLog  *log.Logger
    chunks   *models.ChunkModel
}

// We dont use DefaultServeMux because it is a global variable, 
// any package can access it and register a route — including any third-party
// packages that your application imports. If one of those third-party 
// packages is compromised, they could use DefaultServeMux to expose 
// a malicious handler to the web.

// server mux stores a mapping between the URL patterns for your
// application and the corresponding handlers. The server mux created
// here is a local one, unlike the DefaultServeMux

func main() {
    // Define a new command-line flag with the name 'addr', a default value of ":3001"
    // and some short help text explaining what the flag controls. The value of the
    // flag will be stored in the addr variable at runtime.
    addr := flag.String("addr", ":3001", "HTTP network address")
    // Define a new command-line flag for the MySQL DSN string.
    dsn := flag.String("dsn", "web:pass@/chunkbox?parseTime=true", "MySQL data source name")
    // Importantly, we use the flag.Parse() function to parse the command-line flag.
    // This reads in the command-line flag value and assigns it to the addr
    // variable. You need to call this *before* you use the addr variable
    // otherwise it will always contain the default value of ":3001". If any errors are
    // encountered during parsing the application will be terminated.
    flag.Parse()
    // Use log.New() to create a logger for writing information messages. This takes
    // three parameters: the destination to write the logs to (os.Stdout), a string
    // prefix for message (INFO followed by a tab), and flags to indicate what
    // additional information to include (local date and time). Note that the flags
    // are joined using the bitwise OR operator |.

    infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

    // Create a logger for writing error messages in the same way, but use stderr as
    // the destination and use the log.Lshortfile flag to include the relevant
    // file name and line number.
    errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

    // We pass openDB() the DSN from the command-line flag.
    db, err := openDB(*dsn)
    if err != nil {
        errorLog.Fatal(err)
    }
    // We also defer a call to db.Close(), so that the connection pool is closed
    // before the main() or program exits. It actually will never run
    // in this scenario because of errorLog.Fatal() which terminates
    // the program immediately.
    defer db.Close()
    // Initialize a new instance of our application struct, containing the
    // dependencies.
    app := &application{
        errorLog: errorLog,
        infoLog:  infoLog,
        chunks: &models.ChunkModel{DB:db},
    }
    // Initialize a new http.Server struct. We set the Addr and Handler fields so
    // that the server uses the same network address and routes as before, and set
    // the ErrorLog field so that the server now uses the custom errorLog logger in
    // the event of any problems.
    srv := &http.Server{
        Addr:     *addr,
        ErrorLog: errorLog,
        // call the new app.routes() method to get the servemux containing our routes.
        Handler:  app.routes(),
    }

    // The value returned from the flag.String() function is a pointer to the flag
    // value, not the value itself. So we need to dereference the pointer (i.e.
    // prefix it with the * symbol) before using it. Note that we're using the
    // log.Printf() function to interpolate the address with the log message.
    infoLog.Printf("Starting server on %s", *addr)

    // Instead of the default http.ListenAndServe(), we will use the newly created
    // http server struct. Call the ListenAndServe() method on our new http.Server struct. 
    // err is already declared above.
    err = srv.ListenAndServe()
    errorLog.Fatal(err)
}


// The openDB() function wraps sql.Open() and returns a sql.DB connection pool
// for a given DSN.
func openDB(dsn string) (*sql.DB, error) {
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, err
    }
    //create a connection and check for any errors.
    if err = db.Ping(); err != nil {
        return nil, err
    }
    return db, nil
}