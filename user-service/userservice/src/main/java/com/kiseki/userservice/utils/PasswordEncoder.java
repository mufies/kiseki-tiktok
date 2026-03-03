package com.kiseki.userservice.utils;

import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.stereotype.Component;

/**
 * Thin wrapper around Spring Security's BCryptPasswordEncoder.
 * Kept as its own @Component so AuthService can inject it directly
 * without pulling in the full UserDetailsService contract.
 */
@Component
public class PasswordEncoder {

    private final BCryptPasswordEncoder delegate = new BCryptPasswordEncoder();

    public String encode(String rawPassword) {
        return delegate.encode(rawPassword);
    }

    public boolean matches(String rawPassword, String encodedPassword) {
        return delegate.matches(rawPassword, encodedPassword);
    }
}
