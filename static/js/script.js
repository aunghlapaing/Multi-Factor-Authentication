// Toggle password visibility
function togglePassword(inputId) {
    const passwordInput = document.getElementById(inputId);
    const toggleIcon = passwordInput.nextElementSibling;
    
    if (passwordInput.type === 'password') {
        passwordInput.type = 'text';
        toggleIcon.classList.remove('fa-eye-slash');
        toggleIcon.classList.add('fa-eye');
    } else {
        passwordInput.type = 'password';
        toggleIcon.classList.remove('fa-eye');
        toggleIcon.classList.add('fa-eye-slash');
    }
}

// Password strength checker
function checkPasswordStrength(password) {
    let strength = 0;
    const feedback = {};
    
    // Length check
    if (password.length >= 8) {
        strength += 1;
    } else {
        feedback.length = 'Password should be at least 8 characters long';
    }
    
    // Uppercase check
    if (/[A-Z]/.test(password)) {
        strength += 1;
    } else {
        feedback.uppercase = 'Add uppercase letter';
    }
    
    // Lowercase check
    if (/[a-z]/.test(password)) {
        strength += 1;
    } else {
        feedback.lowercase = 'Add lowercase letter';
    }
    
    // Number check
    if (/[0-9]/.test(password)) {
        strength += 1;
    } else {
        feedback.number = 'Add number';
    }
    
    // Special character check
    if (/[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(password)) {
        strength += 1;
    } else {
        feedback.special = 'Add special character';
    }
    
    return {
        score: strength,
        feedback: feedback
    };
}

// Update password strength meter
function updatePasswordStrength(password) {
    const strengthMeter = document.getElementById('strengthMeter');
    const strengthText = document.getElementById('strengthText');
    
    if (!strengthMeter || !strengthText) return;
    
    const result = checkPasswordStrength(password);
    const score = result.score;
    
    // Update meter
    let meterHTML = '';
    let strengthLabel = '';
    let color = '';
    
    if (score === 0) {
        strengthLabel = 'Very weak';
        color = '#ef4444'; // Red
    } else if (score === 1) {
        strengthLabel = 'Weak';
        color = '#f59e0b'; // Orange
    } else if (score === 2) {
        strengthLabel = 'Fair';
        color = '#f59e0b'; // Orange
    } else if (score === 3) {
        strengthLabel = 'Good';
        color = '#10b981'; // Green
    } else if (score === 4) {
        strengthLabel = 'Strong';
        color = '#10b981'; // Green
    } else if (score === 5) {
        strengthLabel = 'Very strong';
        color = '#10b981'; // Green
    }
    
    meterHTML = `<div style="width: ${score * 20}%; background-color: ${color};"></div>`;
    strengthMeter.innerHTML = meterHTML;
    strengthText.textContent = strengthLabel;
    strengthText.style.color = color;
    
    return result;
}

// Validate email format
function isValidEmail(email) {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
}

// Form validation
document.addEventListener('DOMContentLoaded', function() {
    // Login form validation
    const loginForm = document.getElementById('loginForm');
    if (loginForm) {
        const emailInput = document.getElementById('email');
        const passwordInput = document.getElementById('password');
        const emailError = document.getElementById('emailError');
        const passwordError = document.getElementById('passwordError');
        
        emailInput.addEventListener('blur', function() {
            if (this.value && !isValidEmail(this.value)) {
                emailError.textContent = 'Please enter a valid email address';
            } else {
                emailError.textContent = '';
            }
        });
        
        loginForm.addEventListener('submit', function(e) {
            let isValid = true;
            
            // Validate email
            if (!emailInput.value) {
                emailError.textContent = 'Email is required';
                isValid = false;
            } else if (!isValidEmail(emailInput.value)) {
                emailError.textContent = 'Please enter a valid email address';
                isValid = false;
            } else {
                emailError.textContent = '';
            }
            
            // Validate password
            if (!passwordInput.value) {
                passwordError.textContent = 'Password is required';
                isValid = false;
            } else {
                passwordError.textContent = '';
            }
            
            if (!isValid) {
                e.preventDefault();
            }
        });
    }
    
    // Signup form validation
    const signupForm = document.getElementById('signupForm');
    if (signupForm) {
        const usernameInput = document.getElementById('username');
        const emailInput = document.getElementById('email');
        const passwordInput = document.getElementById('password');
        const confirmPasswordInput = document.getElementById('confirm_password');
        const termsCheckbox = document.getElementById('terms');
        
        const usernameError = document.getElementById('usernameError');
        const emailError = document.getElementById('emailError');
        const passwordError = document.getElementById('passwordError');
        const confirmPasswordError = document.getElementById('confirmPasswordError');
        
        // Real-time password strength check
        if (passwordInput) {
            passwordInput.addEventListener('input', function() {
                const result = updatePasswordStrength(this.value);
                
                // Show feedback
                if (Object.keys(result.feedback).length > 0) {
                    const feedbackList = Object.values(result.feedback).join(', ');
                    passwordError.textContent = feedbackList;
                } else {
                    passwordError.textContent = '';
                }
            });
        }
        
        // Confirm password match check
        if (confirmPasswordInput) {
            confirmPasswordInput.addEventListener('input', function() {
                if (this.value && this.value !== passwordInput.value) {
                    confirmPasswordError.textContent = 'Passwords do not match';
                } else {
                    confirmPasswordError.textContent = '';
                }
            });
        }
        
        // Email validation
        if (emailInput) {
            emailInput.addEventListener('blur', function() {
                if (this.value && !isValidEmail(this.value)) {
                    emailError.textContent = 'Please enter a valid email address';
                } else {
                    emailError.textContent = '';
                }
            });
        }
        
        // Form submission validation
        signupForm.addEventListener('submit', function(e) {
            let isValid = true;
            
            // Validate username
            if (!usernameInput.value) {
                usernameError.textContent = 'Username is required';
                isValid = false;
            } else {
                usernameError.textContent = '';
            }
            
            // Validate email
            if (!emailInput.value) {
                emailError.textContent = 'Email is required';
                isValid = false;
            } else if (!isValidEmail(emailInput.value)) {
                emailError.textContent = 'Please enter a valid email address';
                isValid = false;
            } else {
                emailError.textContent = '';
            }
            
            // Validate password
            if (!passwordInput.value) {
                passwordError.textContent = 'Password is required';
                isValid = false;
            } else {
                const result = checkPasswordStrength(passwordInput.value);
                if (result.score < 3) {
                    passwordError.textContent = 'Password is too weak';
                    isValid = false;
                } else {
                    passwordError.textContent = '';
                }
            }
            
            // Validate confirm password
            if (!confirmPasswordInput.value) {
                confirmPasswordError.textContent = 'Please confirm your password';
                isValid = false;
            } else if (confirmPasswordInput.value !== passwordInput.value) {
                confirmPasswordError.textContent = 'Passwords do not match';
                isValid = false;
            } else {
                confirmPasswordError.textContent = '';
            }
            
            // Validate terms
            if (!termsCheckbox.checked) {
                isValid = false;
                alert('Please agree to the Terms of Service and Privacy Policy');
            }
            
            if (!isValid) {
                e.preventDefault();
            }
        });
    }
    
    // MFA form validation
    const mfaForm = document.getElementById('mfaForm');
    if (mfaForm) {
        const mfaCodeInput = document.getElementById('mfa_code');
        const mfaCodeError = document.getElementById('mfaCodeError');
        
        mfaForm.addEventListener('submit', function(e) {
            let isValid = true;
            
            // Validate MFA code
            if (!mfaCodeInput.value) {
                mfaCodeError.textContent = 'MFA code is required';
                isValid = false;
            } else if (!/^\d{6}$/.test(mfaCodeInput.value)) {
                mfaCodeError.textContent = 'MFA code must be 6 digits';
                isValid = false;
            } else {
                mfaCodeError.textContent = '';
            }
            
            if (!isValid) {
                e.preventDefault();
            }
        });
    }
    
    // MFA setup form validation
    const mfaSetupForm = document.getElementById('mfaSetupForm');
    if (mfaSetupForm) {
        const mfaCodeInput = document.getElementById('mfa_code');
        const mfaCodeError = document.getElementById('mfaCodeError');
        
        mfaSetupForm.addEventListener('submit', function(e) {
            let isValid = true;
            
            // Validate MFA code
            if (!mfaCodeInput.value) {
                mfaCodeError.textContent = 'MFA code is required';
                isValid = false;
            } else if (!/^\d{6}$/.test(mfaCodeInput.value)) {
                mfaCodeError.textContent = 'MFA code must be 6 digits';
                isValid = false;
            } else {
                mfaCodeError.textContent = '';
            }
            
            if (!isValid) {
                e.preventDefault();
            }
        });
    }
});
