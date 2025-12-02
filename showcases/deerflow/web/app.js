document.addEventListener('DOMContentLoaded', () => {
    const queryInput = document.getElementById('queryInput');
    const sendBtn = document.getElementById('sendBtn');
    const messagesContainer = document.getElementById('messages');
    const reportContent = document.getElementById('reportContent');
    const logsContainer = document.getElementById('logsContainer');
    const statusIndicator = document.getElementById('statusIndicator');
    const statusText = statusIndicator.querySelector('.text');
    const tabBtns = document.querySelectorAll('.tab-btn');
    const tabContents = document.querySelectorAll('.tab-content');

    // Tab switching helper
    function switchTab(tabId) {
        // Update buttons
        tabBtns.forEach(b => {
            if (b.dataset.tab === tabId) b.classList.add('active');
            else b.classList.remove('active');
        });

        // Update content
        tabContents.forEach(c => c.classList.remove('active'));
        if (tabId === 'report') {
            document.getElementById('reportContent').classList.add('active');
        } else {
            document.getElementById('activitiesContent').classList.add('active');
        }
    }

    // Tab click handlers
    tabBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            switchTab(btn.dataset.tab);
        });
    });

    // Auto-resize textarea
    queryInput.addEventListener('input', function () {
        this.style.height = 'auto';
        this.style.height = (this.scrollHeight) + 'px';
        sendBtn.disabled = this.value.trim() === '';
    });

    // Handle send
    sendBtn.addEventListener('click', handleSearch);
    queryInput.addEventListener('keydown', (e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            if (!sendBtn.disabled) handleSearch();
        }
    });

    async function handleSearch() {
        const query = queryInput.value.trim();
        if (!query) return;

        // Add user message
        addMessage(query, 'user');
        queryInput.value = '';
        queryInput.style.height = 'auto';
        sendBtn.disabled = true;

        // Set status
        setStatus('Researching...', true);
        switchTab('activities'); // Switch to Activities tab
        reportContent.innerHTML = '<div class="placeholder-text">Initializing research agents...</div>';
        logsContainer.innerHTML = ''; // Clear previous logs

        try {
            // Start SSE connection
            const eventSource = new EventSource(`/api/run?query=${encodeURIComponent(query)}`);

            eventSource.onmessage = (event) => {
                const data = JSON.parse(event.data);

                if (data.type === 'update') {
                    // Update status or partial content
                    if (data.step) {
                        setStatus(`Executing Step: ${data.step}`, true);
                    }
                    if (data.log) {
                        // Optional: Add logs to a console or debug view
                        console.log(data.log);
                    }
                } else if (data.type === 'log') {
                    // Append log
                    const logEntry = document.createElement('div');
                    logEntry.className = 'log-entry';
                    logEntry.textContent = data.message;
                    logsContainer.appendChild(logEntry);
                    logsContainer.scrollTop = logsContainer.scrollHeight;
                } else if (data.type === 'result') {
                    // Final report
                    const report = data.report;
                    reportContent.innerHTML = marked.parse(report);
                    setStatus('Completed', false);
                    switchTab('report'); // Switch back to Report tab
                    eventSource.close();
                } else if (data.type === 'error') {
                    addMessage(`Error: ${data.message}`, 'system');
                    setStatus('Error', false);
                    eventSource.close();
                }
            };

            eventSource.onerror = (err) => {
                console.error('EventSource failed:', err);
                setStatus('Connection Lost', false);
                eventSource.close();
            };

        } catch (error) {
            console.error('Error:', error);
            addMessage('Failed to start research.', 'system');
            setStatus('Error', false);
        }
    }

    function addMessage(text, type) {
        const msgDiv = document.createElement('div');
        msgDiv.className = `message ${type}`;

        let avatarSvg = '';
        if (type === 'user') {
            avatarSvg = `<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>`;
        } else {
            avatarSvg = `<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"></path></svg>`;
        }

        msgDiv.innerHTML = `
            <div class="avatar">${avatarSvg}</div>
            <div class="content"><p>${text}</p></div>
        `;
        messagesContainer.appendChild(msgDiv);
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }

    function setStatus(text, active) {
        statusText.textContent = text;
        if (active) {
            statusIndicator.classList.add('active');
        } else {
            statusIndicator.classList.remove('active');
        }
    }
});
