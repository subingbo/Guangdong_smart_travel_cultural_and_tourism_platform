package com.travelgd.common;

/** Business exception carrying the HTTP status to return. */
public class ApiException extends RuntimeException {
    private final int status;

    public ApiException(int status, String message) {
        super(message);
        this.status = status;
    }

    public static ApiException badRequest(String message) { return new ApiException(400, message); }
    public static ApiException unauthorized(String message) { return new ApiException(401, message); }
    public static ApiException notFound(String message) { return new ApiException(404, message); }
    public static ApiException conflict(String message) { return new ApiException(409, message); }
    public static ApiException badGateway(String message) { return new ApiException(502, message); }

    public int getStatus() { return status; }
}
