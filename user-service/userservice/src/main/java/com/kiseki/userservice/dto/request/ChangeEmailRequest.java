package com.kiseki.userservice.dto.request;

import lombok.Data;

@Data
public class ChangeEmailRequest {
    private String newEmail;
    private String currentPassword;
}
