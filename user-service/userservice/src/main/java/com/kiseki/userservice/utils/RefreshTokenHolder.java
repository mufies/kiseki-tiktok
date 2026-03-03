package com.kiseki.userservice.utils;

/**
 * Thread-local carrier that lets AuthService pass the newly-created
 * refresh token to AuthController without polluting the AuthResponse DTO.
 */
public final class RefreshTokenHolder {

    private static final ThreadLocal<String> HOLDER = new ThreadLocal<>();

    private RefreshTokenHolder() {}

    public static void set(String token) {
        HOLDER.set(token);
    }

    public static String get() {
        return HOLDER.get();
    }

    public static void clear() {
        HOLDER.remove();
    }
}
