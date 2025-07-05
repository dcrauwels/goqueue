console.log('Script loading...');

// Global variables for auth
let authToken = '';
let refreshTokenValue = '';

// Base API URL - adjust this to match your Go server
const API_BASE = 'http://localhost:8080';

console.log('Variables initialized');

// Update auth status display
function updateAuthStatus() {
    console.log('updateAuthStatus called');
    const statusEl = document.getElementById('authStatus');
    if (authToken) {
        statusEl.textContent = 'Authenticated ✓';
        statusEl.className = 'auth-status logged-in';
    } else {
        statusEl.textContent = 'Not authenticated';
        statusEl.className = 'auth-status logged-out';
    }
}

// Generic function to make API requests
async function makeRequest(method, endpoint, data, useAuth) {
    console.log('makeRequest called:', method, endpoint);
    if (useAuth === undefined) useAuth = true;
    
    const url = API_BASE + endpoint;
    const options = {
        method: method,
        headers: {
            'Content-Type': 'application/json',
        },
    };

    if (useAuth && authToken) {
        options.headers['Authorization'] = 'Bearer ' + authToken;
    }

    if (data) {
        options.body = JSON.stringify(data);
    }

    try {
        const response = await fetch(url, options);
        const responseData = await response.text();
        
        let parsedData;
        try {
            parsedData = JSON.parse(responseData);
        } catch (e) {
            parsedData = responseData;
        }

        return {
            status: response.status,
            data: parsedData,
            ok: response.ok
        };
    } catch (error) {
        return {
            status: 0,
            data: 'Network error: ' + error.message,
            ok: false
        };
    }
}

// Display response in a container
function displayResponse(elementId, response) {
    console.log('displayResponse called:', elementId);
    const element = document.getElementById(elementId);
    element.style.display = 'block';
    element.className = 'response ' + (response.ok ? 'success' : 'error');
    element.textContent = 'Status: ' + response.status + '\n\n' + JSON.stringify(response.data, null, 2);
}

// Status functions
function checkHealth() {
    console.log('checkHealth called');
    makeRequest('GET', '/api/healthz', null, false).then(function(response) {
        displayResponse('healthResponse', response);
    });
}

// Auth functions
function login() {
    console.log('login called');
    const email = document.getElementById('loginEmail').value;
    const password = document.getElementById('loginPassword').value;
    
    if (!email || !password) {
        alert('Please enter both username and password');
        return;
    }

    const loginData = { email: email, password: password };
    makeRequest('POST', '/api/login', loginData, false).then(function(response) {
        if (response.ok && response.data.user_access_token) {
            authToken = response.data.user_access_token;
            if (response.data.user_refresh_token) {
                refreshTokenValue = response.data.user_refresh_token;
            }
            updateAuthStatus();
            setCookie('accessToken', authToken, 1)
        }
        
        displayResponse('loginResponse', response);
    });
}

function refreshTokenFunc() {
    console.log('refreshTokenFunc called');
    makeRequest('POST', '/api/refresh', { refresh_token: refreshTokenValue, user_id: userIDValue }, false).then(function(response) {
        if (response.ok && response.data.user_access_token) {
            authToken = response.data.token;
            updateAuthStatus();
        }
        displayResponse('sessionResponse', response);
    });
}

function logout() {
    console.log('logout called');
    makeRequest('POST', '/api/logout').then(function(response) {
        authToken = '';
        refreshTokenValue = '';
        updateAuthStatus();
        displayResponse('sessionResponse', response);
    });
}

// User functions
function createUser() {
    console.log('createUser called');
    const dataText = document.getElementById('newUserData').value;
    try {
        const userData = JSON.parse(dataText);
        makeRequest('POST', '/api/users', userData).then(function(response) {
            displayResponse('createUserResponse', response);
        });
    } catch (error) {
        displayResponse('createUserResponse', {
            status: 0,
            data: 'Invalid JSON: ' + error.message,
            ok: false
        });
    }
}

function getUsers() {
    console.log('getUsers called');
    const userId = document.getElementById('getUserId').value;
    const endpoint = userId ? '/api/users/' + userId : '/api/users';
    makeRequest('GET', endpoint).then(function(response) {
        displayResponse('getUsersResponse', response);
    });
}

function updateUser() {
    console.log('updateUser called');
    const userId = document.getElementById('updateUserId').value;
    const dataText = document.getElementById('updateUserData').value;
    
    if (!userId) {
        alert('Please enter a user ID');
        return;
    }

    try {
        const userData = JSON.parse(dataText);
        makeRequest('PUT', '/api/users/' + userId, userData).then(function(response) {
            displayResponse('updateUserResponse', response);
        });
    } catch (error) {
        displayResponse('updateUserResponse', {
            status: 0,
            data: 'Invalid JSON: ' + error.message,
            ok: false
        });
    }
}

function updateUserSelf() {
    console.log('updateUserSelf called');
    const dataText = document.getElementById('updateUserData').value;
    try {
        const userData = JSON.parse(dataText);
        makeRequest('PUT', '/api/users', userData).then(function(response) {
            displayResponse('updateUserResponse', response);
        });
    } catch (error) {
        displayResponse('updateUserResponse', {
            status: 0,
            data: 'Invalid JSON: ' + error.message,
            ok: false
        });
    }
}

// Purpose functions
function createPurpose() {
    console.log('createPurpose called');
    const dataText = document.getElementById('newPurposeData').value;
    try {
        const purposeData = JSON.parse(dataText);
        makeRequest('POST', '/api/purposes', purposeData).then(function(response) {
            displayResponse('createPurposeResponse', response);
        });
    } catch (error) {
        displayResponse('createPurposeResponse', {
            status: 0,
            data: 'Invalid JSON: ' + error.message,
            ok: false
        });
    }
}

function getPurposes() {    
    console.log('getPurposes called');
    const purposeId = document.getElementById('getPurposeId').value;
    const endpoint = purposeId ? '/api/purposes/' + purposeId : '/api/purposes';
    makeRequest('GET', endpoint).then(function(response) {
        displayResponse('getPurposesResponse', response);
    });
}

function updatePurpose() {
    console.log('updatePurpose called');
    const purposeId = document.getElementById('updatePurposeId').value;
    const dataText = document.getElementById('updatePurposeData').value;
    
    if (!purposeId) {
        alert('Please enter a purpose ID');
        return;
    }

    try {
        const purposeData = JSON.parse(dataText);
        makeRequest('PUT', '/api/purposes/' + purposeId, purposeData).then(function(response) {
            displayResponse('updatePurposeResponse', response);
        });
    } catch (error) {
        displayResponse('updatePurposeResponse', {
            status: 0,
            data: 'Invalid JSON: ' + error.message,
            ok: false
        });
    }
}


// Visitor functions
function createVisitor() {
    console.log('createVisitor called');
    const dataText = document.getElementById('newVisitorData').value;
    try {
        const visitorData = JSON.parse(dataText);
        makeRequest('POST', '/api/visitors', visitorData).then(function(response) {
            displayResponse('createVisitorResponse', response);
        });
    } catch (error) {
        displayResponse('createVisitorResponse', {
            status: 0,
            data: 'Invalid JSON: ' + error.message,
            ok: false
        });
    }
}

function getVisitors() {
    console.log('getVisitors called');
    const visitorId = document.getElementById('getVisitorId').value;
    const endpoint = visitorId ? '/api/visitors/' + visitorId : '/api/visitors';
    makeRequest('GET', endpoint).then(function(response) {
        displayResponse('getVisitorsResponse', response);
    });
}

function updateVisitor() {
    console.log('updateVisitor called');
    const visitorId = document.getElementById('updateVisitorId').value;
    const dataText = document.getElementById('updateVisitorData').value;
    
    if (!visitorId) {
        alert('Please enter a visitor ID');
        return;
    }

    try {
        const visitorData = JSON.parse(dataText);
        makeRequest('PUT', '/api/visitors/' + visitorId, visitorData).then(function(response) {
            displayResponse('updateVisitorResponse', response);
        });
    } catch (error) {
        displayResponse('updateVisitorResponse', {
            status: 0,
            data: 'Invalid JSON: ' + error.message,
            ok: false
        });
    }
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    console.log('DOM loaded, initializing...');
    authToken = getCookie('accessToken');
    updateAuthStatus();
    console.log('Script loaded successfully');
});

// cookie handling

// distribute cookie
function setCookie(name, value, hours) {
    console.log('setCookie called')
    let expires = '';
    if (days) {
        const date = new Date();
        date.setTime(date.getTime() + (hours * 60 * 60 * 1000));
        expires = '; expires=' + date.toUTCString();
    }
    document.cookie = name + '=' + (value || '') + expires + '; path=/; SameSite=Lax';
    // Consider adding 'Secure;' in production if using HTTPS
    // document.cookie = name + '=' + (value || '') + expires + '; path=/; SameSite=Lax; Secure';
}

// retrieve cookie
function getCookie(name) {
    console.log('getCookie called with name ' + name)
    const nameEQ = name + '=';
    const ca = document.cookie.split(';');
    for (let i = 0; i < ca.length; i++) {
        let c = ca[i];
        while (c.charAt(0) === ' ') c = c.substring(1, c.length);
        if (c.indexOf(nameEQ) === 0) return c.substring(nameEQ.length, c.length);
    }
    return null;
}