package com.kiseki.userservice.dto.request;

import lombok.Data;

@Data
public class LoginRequest {
    private String email;
    private String username;
    private String password;
}

