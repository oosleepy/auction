const DOM = {
    authSection: document.getElementById('auth-section'),
    dashboardSection: document.getElementById('dashboard-section'),
    tabLogin: document.getElementById('tab-login'),
    tabRegister: document.getElementById('tab-register'),
    authForm: document.getElementById('auth-form'),
    authSubmit: document.getElementById('auth-submit'),
    authMsg: document.getElementById('auth-msg'),
    
    userDisplay: document.getElementById('user-display'),
    totalBid: document.getElementById('total-bid'),
    logoutBtn: document.getElementById('logout-btn'),
    
    createAuctionForm: document.getElementById('create-auction-form'),
    createMsg: document.getElementById('create-msg'),
    
    activeList: document.getElementById('active-list'),
    refreshActive: document.getElementById('refresh-active'),
    
    historyList: document.getElementById('history-list'),
    refreshHistory: document.getElementById('refresh-history'),
    
    biddingModal: document.getElementById('bidding-modal'),
    modalAuctionName: document.getElementById('modal-auction-name'),
    modalCurrentBid: document.getElementById('modal-current-bid'),
    placeBidForm: document.getElementById('place-bid-form'),
    newBid: document.getElementById('new-bid'),
    bidMsg: document.getElementById('bid-msg'),
    closeBtn: document.querySelector('.close-btn')
};

let currentMode = 'login'; // 'login' or 'register'
let currentWs = null;
let activeModalAuction = null;

// Initialize
function init() {
    checkAuth();
    attachListeners();
}

function attachListeners() {
    DOM.tabLogin.addEventListener('click', () => switchAuthMode('login'));
    DOM.tabRegister.addEventListener('click', () => switchAuthMode('register'));
    DOM.authForm.addEventListener('submit', handleAuth);
    DOM.logoutBtn.addEventListener('click', logout);
    DOM.createAuctionForm.addEventListener('submit', handleCreateAuction);
    DOM.refreshActive.addEventListener('click', loadActiveAuctions);
    DOM.refreshHistory.addEventListener('click', loadHistory);
    DOM.closeBtn.addEventListener('click', closeModal);
    DOM.placeBidForm.addEventListener('submit', handlePlaceBid);
}

// Authentication
function switchAuthMode(mode) {
    currentMode = mode;
    if (mode === 'login') {
        DOM.tabLogin.classList.add('active');
        DOM.tabRegister.classList.remove('active');
        DOM.authSubmit.textContent = 'Login';
    } else {
        DOM.tabRegister.classList.add('active');
        DOM.tabLogin.classList.remove('active');
        DOM.authSubmit.textContent = 'Register';
    }
    DOM.authMsg.textContent = '';
}

function checkAuth() {
    const token = localStorage.getItem('token');
    if (token) {
        showDashboard();
        loadActiveAuctions();
        loadHistory();
        loadTotalBid();
    } else {
        showAuth();
    }
}

function showAuth() {
    DOM.authSection.classList.remove('hidden');
    DOM.dashboardSection.classList.add('hidden');
    DOM.biddingModal.classList.add('hidden');
}

function showDashboard() {
    DOM.authSection.classList.add('hidden');
    DOM.dashboardSection.classList.remove('hidden');
    DOM.userDisplay.textContent = localStorage.getItem('username') || 'User';
}

function logout() {
    localStorage.removeItem('token');
    localStorage.removeItem('username');
    if (currentWs) {
        currentWs.close();
    }
    showAuth();
}

async function handleAuth(e) {
    e.preventDefault();
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const endpoint = currentMode === 'login' ? '/login' : '/register';
    
    DOM.authMsg.textContent = 'Processing...';
    DOM.authMsg.className = 'msg';
    
    try {
        const res = await fetch(endpoint, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });
        
        if (res.ok) {
            if (currentMode === 'login') {
                const data = await res.json();
                localStorage.setItem('token', data.token);
                localStorage.setItem('username', username);
                checkAuth();
            } else {
                DOM.authMsg.textContent = 'Registration successful. Please login.';
                DOM.authMsg.className = 'msg success';
                switchAuthMode('login');
            }
        } else {
            const err = await res.text();
            DOM.authMsg.textContent = err || 'Authentication failed';
            DOM.authMsg.className = 'msg error';
        }
    } catch (err) {
        DOM.authMsg.textContent = 'Network error';
        DOM.authMsg.className = 'msg error';
    }
}

function getAuthHeaders() {
    return {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + localStorage.getItem('token')
    };
}

// User Info
async function loadTotalBid() {
    try {
        const res = await fetch('/mybid', { headers: getAuthHeaders() });
        if (res.ok) {
            const data = await res.json();
            DOM.totalBid.textContent = data.totalbid || 0;
        }
    } catch (e) {
        console.error("Could not load total bid", e);
    }
}

// Dashboard Actions
async function handleCreateAuction(e) {
    e.preventDefault();
    const name = document.getElementById('auction-name').value;
    const bid = document.getElementById('initial-bid').value;
    const expiry = parseInt(document.getElementById('expiry-time').value);
    
    DOM.createMsg.textContent = 'Creating...';
    DOM.createMsg.className = 'msg';
    
    try {
        const res = await fetch('/setbid', {
            method: 'POST',
            headers: getAuthHeaders(),
            body: JSON.stringify({ name, bid, time: expiry })
        });
        
        if (res.ok) {
            DOM.createMsg.textContent = 'Auction created!';
            DOM.createMsg.className = 'msg success';
            DOM.createAuctionForm.reset();
            loadActiveAuctions();
        } else {
            const err = await res.text();
            DOM.createMsg.textContent = err || 'Failed to create';
            DOM.createMsg.className = 'msg error';
        }
    } catch (err) {
        DOM.createMsg.textContent = 'Network error';
        DOM.createMsg.className = 'msg error';
    }
}

async function loadActiveAuctions() {
    try {
        const res = await fetch('/listactive');
        if (res.ok) {
            const data = await res.json();
            const list = data.active_auciton || []; // Matches JSON tag in Go struct
            
            DOM.activeList.innerHTML = '';
            if (list.length === 0) {
                DOM.activeList.innerHTML = '<li><p>No active auctions.</p></li>';
                return;
            }
            
            for (const name of list) {
                // Fetch current bid for each to display
                const bidRes = await fetch('/getbid', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ name })
                });
                
                let currentBid = '?';
                if (bidRes.ok) {
                    const bidData = await bidRes.json();
                    currentBid = bidData.bid;
                }
                
                const li = document.createElement('li');
                li.innerHTML = `
                    <div class="item-info">
                        <h4>${name}</h4>
                        <p>Current Bid: <strong>${currentBid}</strong></p>
                    </div>
                    <div class="item-action">
                        <button onclick="openModal('${name}', '${currentBid}')">Join</button>
                    </div>
                `;
                DOM.activeList.appendChild(li);
            }
        }
    } catch (e) {
        console.error(e);
        DOM.activeList.innerHTML = '<li><p class="error">Failed to load</p></li>';
    }
}

async function loadHistory() {
    try {
        const res = await fetch('/listhistory');
        if (res.ok) {
            const data = await res.json();
            DOM.historyList.innerHTML = '';
            if (!data || data.length === 0) {
                DOM.historyList.innerHTML = '<li><p>No history found.</p></li>';
                return;
            }
            
            data.forEach(item => {
                const date = new Date(item.created_at).toLocaleString();
                const li = document.createElement('li');
                li.innerHTML = `
                    <div class="item-info">
                        <h4>${item.bid_name}</h4>
                        <p>Winner ID: ${item.userid} | Winning Bid: <strong>${item.bid_amt}</strong></p>
                        <p style="font-size: 0.75rem; color: #999;">${date}</p>
                    </div>
                `;
                DOM.historyList.appendChild(li);
            });
        }
    } catch (e) {
        console.error(e);
        DOM.historyList.innerHTML = '<li><p class="error">Failed to load history</p></li>';
    }
}

// Bidding Modal and WS
function openModal(name, initialBid) {
    activeModalAuction = name;
    DOM.modalAuctionName.textContent = name;
    DOM.modalCurrentBid.textContent = initialBid;
    DOM.bidMsg.textContent = '';
    DOM.newBid.value = '';
    DOM.biddingModal.classList.remove('hidden');
    
    connectWS(name);
}

function closeModal() {
    DOM.biddingModal.classList.add('hidden');
    if (currentWs) {
        currentWs.close();
        currentWs = null;
    }
    activeModalAuction = null;
    loadActiveAuctions();
}

function connectWS(name) {
    if (currentWs) currentWs.close();
    
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws?name=${name}`;
    
    currentWs = new WebSocket(wsUrl);
    
    currentWs.onmessage = (event) => {
        const data = JSON.parse(event.data);
        if (data.name === activeModalAuction) {
            DOM.modalCurrentBid.textContent = data.bid;
            // Add a little pop effect
            DOM.modalCurrentBid.style.transform = 'scale(1.2)';
            setTimeout(() => {
                DOM.modalCurrentBid.style.transform = 'scale(1)';
            }, 200);
        }
    };
    
    currentWs.onerror = (error) => {
        console.error("WS Error:", error);
    };
}

async function handlePlaceBid(e) {
    e.preventDefault();
    if (!activeModalAuction) return;
    
    const bidVal = DOM.newBid.value;
    DOM.bidMsg.textContent = 'Placing...';
    DOM.bidMsg.className = 'msg';
    
    try {
        const res = await fetch('/bid', {
            method: 'POST',
            headers: getAuthHeaders(),
            body: JSON.stringify({ name: activeModalAuction, bid: bidVal })
        });
        
        if (res.status === 202) {
            DOM.bidMsg.textContent = 'Bid accepted!';
            DOM.bidMsg.className = 'msg success';
            DOM.newBid.value = '';
            loadTotalBid();
        } else if (res.status === 409) {
            DOM.bidMsg.textContent = 'Bid too low';
            DOM.bidMsg.className = 'msg error';
        } else {
            DOM.bidMsg.textContent = 'Failed to place bid';
            DOM.bidMsg.className = 'msg error';
        }
    } catch (err) {
        DOM.bidMsg.textContent = 'Network error';
        DOM.bidMsg.className = 'msg error';
    }
}

// Ensure smooth transitions
DOM.modalCurrentBid.style.transition = 'transform 0.2s cubic-bezier(0.175, 0.885, 0.32, 1.275)';

init();
