document.addEventListener('DOMContentLoaded', function() {
    const loginForm = document.getElementById('loginForm');
    const submitBtn = document.getElementById('submitBtn');
    const usernameInput = document.getElementById('username');
    const passwordInput = document.getElementById('password');

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

    loginForm.addEventListener('submit', async function(e) {
        e.preventDefault();
        
        const username = usernameInput.value.trim();
        const password = passwordInput.value.trim();

        if (!username || !password) {
            showToast('Validation Error', 'Please fill in all fields', 'destructive');
            return;
        }

        submitBtn.disabled = true;
        submitBtn.textContent = 'Logging in...';

        try {
            const response = await fetch('/api/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username, password })
            });
            
            const data = await response.json();
            
            if (!response.ok || !data.success) {
                let errorMessage = 'Invalid credentials. Please check your username and password.';
                if (response.status === 401) {
                    errorMessage = 'Invalid username or password. If you don\'t have an account, please register.';
                }
                throw new Error(errorMessage);
            }
            
            localStorage.setItem('authToken', data.token);
            localStorage.setItem('username', username);
            
            showToast('Login Successful', 'Welcome back!');
            
            setTimeout(() => {
                window.location.href = 'dashboard.html';
            }, 500);
        } catch (error) {
            showToast('Login Failed', error.message || 'Invalid credentials. Please try again.', 'destructive');
            submitBtn.disabled = false;
            submitBtn.textContent = 'Sign In';
        }
    });
});
