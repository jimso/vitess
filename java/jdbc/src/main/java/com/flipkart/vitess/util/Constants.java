package com.flipkart.vitess.util;

/**
 * Created by harshit.gangal on 25/01/16.
 */
public class Constants {

    public static final boolean JDBC_COMPLIANT = false;
    public static final String URL_PREFIX = "jdbc:vitess://";
    public static final String URL_PATTERN =
        "(jdbc:vitess:\\/\\/)([a-zA-Z0-9.-]+)(:([0-9]*))*(\\/([a-zA-Z0-9_$]*|))(\\/([a-zA-Z0-9_$]*|))*";
    public static final String VITESS_HOST = "Hostname of Vitess Server";
    public static final String VITESS_PORT = "Port number of Vitess Server";
    public static final String VITESS_DB_NAME = "Database name";
    public static final String VITESS_TABLET_TYPE =
        "Tablet Type to which Vitess will connect(master, replica, rdonly)";
    public static final String DEFAULT_HOST = "localhost";
    public static final String DEFAULT_PORT = "15991";
    public static final String DEFAULT_DBNAME = "";
    public static final String DEFAULT_TABLET_TYPE = "MASTER";
    public static final long CONNECTION_TIMEOUT = 30000;
    public static final String LITERAL_V = "v";
    public static final String LITERAL_SINGLE_QUOTE = "'";
    public static final String SQL_SELECT = "select";
    public static final String SQL_INSERT = "insert";
    public static final String SQL_UPDATE = "update";
    public static final String SQL_DELETE = "delete";
    public static final int DRIVER_MAJOR_VERSION = 1;
    public static final int DRIVER_MINOR_VERSION = 0;
    public static final int MAX_BUFFER_SIZE = 65535;
    public static final int DEFAULT_TIMEOUT = 30000; //In miliseconds
    public static final String LITERAL_S = "s";
    public static final String SQL_SHOW = "show";
    public static final String VITESS_KEYSPACE = "Keyspace name in Vitess Server";
    public static final Constants.QueryExecuteType DEFAULT_EXECUTE_TYPE = QueryExecuteType.SIMPLE;
    public static final String EXECUTE_TYPE_DESC = "Query execution type: simple or stream \n";
    public static final String USERNAME_DESC = "Username used for ACL validation \n";


    private Constants() {
    }


    public static final class SQLExceptionMessages {
        public static final String CONN_CLOSED = "Connection is Closed";
        public static final String ERROR_CLOSE = "Error while cleaning up the server resources";
        public static final String INIT_FAILED = "Failed to Initialize Vitess JDBC Driver";
        public static final String INVALID_CONN_URL = "Connection URL is invalid";
        public static final String READ_ONLY = "Read Only mode is not supported";
        public static final String STMT_CLOSED = "Statement is closed";
        public static final String SQL_FEATURE_NOT_SUPPORTED = "SQL Feature Not Supported";
        public static final String TX_CLOSED = "No Transaction Exists";
        public static final String TIMEOUT_NEGATIVE = "Timeout value cannot be negative";
        public static final String COMMIT_WHEN_AUTO_COMMIT_TRUE =
            "Cannot call commit when auto commit is true";
        public static final String ROLLBACK_WHEN_AUTO_COMMIT_TRUE =
            "Cannot call commit when auto commit is true";
        public static final String CLOSED_RESULT_SET = "Result Set closed";
        public static final String INVALID_COLUMN_INDEX = "Invalid Column Index";
        public static final String NO_CURRENT_ROW = "Invalid Column Index";
        public static final String VITESS_CURSOR_CLOSE_ERROR =
            "Getting Error while closing ResultSet";
        public static final String CONN_INIT_ERROR = "Connection initialization error";
        public static final String BAD_FORMAT_BIG_DECIMAL = "Bad format for big decimal";
        public static final String COLUMN_NOT_FOUND = "column not found";
        public static final String MALFORMED_URL = "Malformed URL Exception";
        public static final String SQL_TYPE_INFER =
            "Cannot infer the SQL type to use for an instance of ";
        public static final String DML_NOT_ON_MASTER =
            "DML Statement cannot be executed on non master instance type";
        public static final String SQL_NOT_SUPPORTED = "This SQL statement is not supported";
        public static final String SQL_EMPTY = "SQL statement is not valid";
        public static final String RESULT_SET_TYPE_NOT_SUPPORTED =
            "This Result Set type is not supported";
        public static final String RESULT_SET_CONCUR_NOT_SUPPORTED =
            "This Result Set Concurrency is not supported";
        public static final String METHOD_CALLED_ON_OPEN_TRANSACTION =
            "This method should not be called when a transaction is open";
        public static final String ISOLATION_LEVEL_NOT_SUPPORTED =
            "This isolation level is not supported";
        public static final String EXECUTOR_NULL = "Executor cannot be null";
        public static final String CLASS_CAST_EXCEPTION = "Unable to unwrap to ";
        public static final String INVALID_COLUMN_TYPE = "Invalid Column Type";
        public static final String UNKNOWN_COLUMN_TYPE = "Unknown Column Type";
        public static final String INVALID_RESULT_SET = "Unable to build ResultSet";
        public static final String METHOD_NOT_ALLOWED =
            "This method cannot be called using this class object";
        public static final String SQL_RETURNED_RESULT_SET =
            "ResultSet generation is not allowed through this method";
        public static final String ILLEGAL_VALUE_FOR = "Illegal value for ";
        public static final String METHOD_CALL_FAILED = "Failed to execute this method";
        public static final String KEYSPACE_REQUIRED =
            "Keyspace name is required in the connection url";
        public static final String CURSOR_NULL = "Cursor cannot be null";
        public static final String NO_COLUMN_ACCESSED =
            "No column was accessed before calling this method";
        public static final String DBNAME_REQUIRED =
            "Database name is required in the connection url";
        public static final String RESULT_SET_INIT_ERROR = "ResultSet initialization error";
    }


    public static final class Property {
        public static final String TABLET_TYPE = "TABLET_TYPE";
        public static final String HOST = "HOST";
        public static final String PORT = "PORT";
        public static final String DBNAME = "DBNAME";
        public static final String KEYSPACE = "KEYSPACE";
        public static final String USERNAME = "userName";
        public static final String EXECUTE_TYPE = "executeType";
    }

    public enum QueryExecuteType {
        SIMPLE,
        STREAM
    }
}
