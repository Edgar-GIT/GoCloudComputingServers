document.addEventListener('DOMContentLoaded', function() {
    const registerForm = document.getElementById('registerForm');
    const submitBtn = document.getElementById('submitBtn');
    const usernameInput = document.getElementById('username');
    const passwordInput = document.getElementById('password');
    const confirmPasswordInput = document.getElementById('confirmPassword');

    function showToast(title, description, variant = 'default') {
        const toastContainer = document.getElementById('toastContainer');
        const toast = document.createElement('div');
        toast.className = `toast ${variant === 'destructive' ? 'destructive' : ''}`;
        
        const titleEl = document.createElement('div');
        titleEl.style.fontWeight = '600';
        titleEl.style.marginBottom = '0.25rem';
        titleEl.textContent = title;
        
        const descEl = document.createElement('div');
        descEl.style.fontSize = '0.875rem';
        descEl.style.color = 'hsl(var(--muted-foreground))';
        descEl.textContent = description;
        
        toast.appendChild(titleEl);
        toast.appendChild(descEl);
        toastContainer.appendChild(toast);

        setTimeout(() => {
            toast.style.animation = 'slide-in 0.3s ease-out reverse';
            setTimeout(() => toast.remove(), 300);
        }, 3000);
    }

    function validateUsername(username) {
        const usernameRegex = /^[a-zA-Z0-9_]+$/;
        return usernameRegex.test(username);
    }

    registerForm.addEventListener('submit', async function(e) {
        e.preventDefault();
        
        const username = usernameInput.value.trim();
        const password = passwordInput.value.trim();
        const confirmPassword = confirmPasswordInput.value.trim();

        if (!username || !password || !confirmPassword) {
            showToast('Validation Error', 'Please fill in all fields', 'destructive');
            return;
        }

        if (!validateUsername(username)) {
            showToast('Invalid Username', 'Username can only contain letters, numbers, and underscores', 'destructive');
            return;
        }

        if (username.length < 3) {
            showToast('Invalid Username', 'Username must be at least 3 characters long', 'destructive');
            return;
        }

        if (username.toLowerCase() === 'admin') {
            showToast('Invalid Username', 'The username "admin" is reserved', 'destructive');
            return;
        }

        if (password.length < 3) {
            showToast('Invalid Password', 'Password must be at least 3 characters long', 'destructive');
            return;
        }

        if (password !== confirmPassword) {
            showToast('Password Mismatch', 'Passwords do not match', 'destructive');
            return;
        }

        submitBtn.disabled = true;
        submitBtn.textContent = 'Creating account...';

        try {
            const response = await fetch('/api/register', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username, password })
            });
            
            const data = await response.json();
            
            if (!response.ok) {
                let errorMessage = 'Failed to create account. Please try again.';
                if (data.error) {
                    errorMessage = data.error;
                } else if (response.status === 400) {
                    errorMessage = 'Invalid username or password. Please check your input.';
                }
                throw new Error(errorMessage);
            }
            
            showToast('Account Created', 'Your account has been created successfully!');
            
            setTimeout(() => {
                window.location.href = 'login.html';
            }, 1500);
        } catch (error) {
            showToast('Registration Failed', error.message || 'Failed to create account. Please try again.', 'destructive');
            submitBtn.disabled = false;
            submitBtn.textContent = 'Create Account';
        }
    });

    confirmPasswordInput.addEventListener('input', function() {
        if (confirmPasswordInput.value && passwordInput.value) {
            if (confirmPasswordInput.value !== passwordInput.value) {
                confirmPasswordInput.setCustomValidity('Passwords do not match');
            } else {
                confirmPasswordInput.setCustomValidity('');
            }
        }
    });

    passwordInput.addEventListener('input', function() {
        if (confirmPasswordInput.value && passwordInput.value) {
            if (confirmPasswordInput.value !== passwordInput.value) {
                confirmPasswordInput.setCustomValidity('Passwords do not match');
            } else {
                confirmPasswordInput.setCustomValidity('');
            }
        }
    });
});
