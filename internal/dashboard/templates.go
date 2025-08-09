package dashboard

// Templates contains all HTML templates for the dashboard
const (
	MainTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>LED Screen Contracts Dashboard</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background-color: #000000;
            color: #ffffff;
            min-height: 100vh;
        }
        
        /* Top green line */
        body::before {
            content: '';
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            height: 2px;
            background-color: #00ff00;
            z-index: 1000;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            min-height: 100vh;
        }
        
        .header {
            text-align: center;
            margin-bottom: 30px;
        }
        
        .logo {
            font-size: 2.5em;
            font-weight: bold;
            margin-bottom: 10px;
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 10px;
        }
        
        .logo-symbol {
            color: #ffffff;
        }
        
        .logo-text {
            color: #ffffff;
        }
        
        .title {
            font-size: 1.8em;
            color: #ffffff;
            margin-bottom: 20px;
        }
        
        .stats {
            display: flex;
            justify-content: space-around;
            padding: 20px;
            background: #1a1a1a;
            border-radius: 8px;
            margin-bottom: 20px;
            border: 1px solid #333333;
        }
        
        .stat {
            text-align: center;
        }
        
        .stat-number {
            font-size: 2.5em;
            font-weight: bold;
            color: #ff6600;
        }
        
        .stat-label {
            color: #ffffff;
            font-size: 0.9em;
            margin-top: 5px;
        }
        
        .controls {
            padding: 20px;
            background: #1a1a1a;
            border-radius: 8px;
            margin-bottom: 20px;
            border: 1px solid #333333;
            display: flex;
            gap: 15px;
            align-items: center;
        }
        
        .btn {
            padding: 12px 24px;
            border: none;
            border-radius: 6px;
            cursor: pointer;
            font-size: 14px;
            font-weight: 500;
            transition: all 0.3s ease;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        
        .btn-primary {
            background: #ff6600;
            color: white;
        }
        
        .btn-primary:hover {
            background: #e55a00;
            transform: translateY(-1px);
            box-shadow: 0 4px 12px rgba(255, 102, 0, 0.3);
        }
        
        .btn-danger {
            background: #ff3333;
            color: white;
        }
        
        .btn-danger:hover {
            background: #e60000;
            transform: translateY(-1px);
            box-shadow: 0 4px 12px rgba(255, 51, 51, 0.3);
        }
        
        .search {
            flex: 1;
            padding: 12px 16px;
            border: 1px solid #333333;
            border-radius: 6px;
            font-size: 14px;
            background: #000000;
            color: #ffffff;
            transition: all 0.3s ease;
        }
        
        .search:focus {
            outline: none;
            border-color: #ff6600;
            box-shadow: 0 0 0 2px rgba(255, 102, 0, 0.2);
        }
        
        .search::placeholder {
            color: #666666;
        }
        
        .contracts {
            padding: 20px 0;
        }
        
        .contract {
            border: 1px solid #333333;
            border-radius: 8px;
            margin-bottom: 20px;
            overflow: hidden;
            transition: all 0.3s ease;
            background: #1a1a1a;
        }
        
        .contract:hover {
            box-shadow: 0 8px 25px rgba(255, 102, 0, 0.15);
            transform: translateY(-3px);
            border-color: #ff6600;
        }
        
        .contract-header {
            background: #2a2a2a;
            padding: 20px;
            border-bottom: 1px solid #333333;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        .contract-actions {
            display: flex;
            align-items: center;
            gap: 10px;
        }
        
        .delete-contract-btn {
            background: #ff3333;
            color: #ffffff;
            border: none;
            border-radius: 50%;
            width: 32px;
            height: 32px;
            cursor: pointer;
            font-size: 20px;
            font-weight: bold;
            display: flex;
            align-items: center;
            justify-content: center;
            transition: all 0.3s ease;
            line-height: 1;
        }
        
        .delete-contract-btn:hover {
            background: #cc0000;
            transform: scale(1.1);
        }
        
        .contract-id {
            font-weight: bold;
            color: #ff6600;
            font-size: 1.2em;
        }
        
        .contract-status {
            padding: 6px 16px;
            border-radius: 20px;
            font-size: 0.8em;
            font-weight: bold;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        
        .status-publicada {
            background: #00ff00;
            color: #000000;
        }
        
        .status-adjudicada {
            background: #ff6600;
            color: #ffffff;
        }
        
        .status-anulada {
            background: #ff3333;
            color: #ffffff;
        }
        
        .status-evaluación-previa {
            background: linear-gradient(135deg, #ff6600, #ff9933);
            color: #ffffff;
            box-shadow: 0 4px 15px rgba(255, 102, 0, 0.3);
            border: 1px solid #ff6600;
            animation: pulse 2s infinite;
        }
        
        @keyframes pulse {
            0% {
                box-shadow: 0 4px 15px rgba(255, 102, 0, 0.3);
            }
            50% {
                box-shadow: 0 4px 20px rgba(255, 102, 0, 0.5);
            }
            100% {
                box-shadow: 0 4px 15px rgba(255, 102, 0, 0.3);
            }
        }
        
        .contract-body {
            padding: 25px;
        }
        
        .contract-description {
            font-size: 1.1em;
            margin-bottom: 20px;
            line-height: 1.6;
            color: #ffffff;
        }
        
        .contract-details {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            font-size: 0.9em;
        }
        
        .detail-item {
            display: flex;
            flex-direction: column;
            padding: 15px;
            background: #000000;
            border-radius: 6px;
            border: 1px solid #333333;
        }
        
        .detail-label {
            font-weight: bold;
            color: #ff6600;
            margin-bottom: 8px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            font-size: 0.8em;
        }
        
        .detail-item > div:last-child {
            color: #ffffff;
        }
        
        .amount {
            color: #00ff00;
            font-weight: bold;
            font-size: 1.1em;
        }
        
        .status-changes {
            background: #1a1a1a;
            border-radius: 8px;
            margin-bottom: 20px;
            border: 1px solid #333333;
            padding: 20px;
        }
        
        .status-change-item {
            background: #000000;
            border-radius: 6px;
            padding: 15px;
            margin-bottom: 10px;
            border: 1px solid #333333;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        .status-change-info {
            flex: 1;
        }
        
        .status-change-contract {
            color: #ff6600;
            font-weight: bold;
            font-size: 1.1em;
            margin-bottom: 5px;
        }
        
        .status-change-details {
            color: #ffffff;
            font-size: 0.9em;
        }
        
        .status-change-arrow {
            color: #ff6600;
            font-weight: bold;
            margin: 0 10px;
        }
        
        .status-change-time {
            color: #666666;
            font-size: 0.8em;
            text-align: right;
        }
        
        .status-change-checkmark {
            background: #00ff00;
            color: #000000;
            border: none;
            border-radius: 50%;
            width: 24px;
            height: 24px;
            cursor: pointer;
            font-size: 12px;
            display: flex;
            align-items: center;
            justify-content: center;
            transition: all 0.3s ease;
            margin-left: 10px;
        }
        
        .status-change-checkmark:hover {
            background: #00cc00;
            transform: scale(1.1);
        }
        
        .status-change-item.vanishing {
            animation: vanish 0.5s ease-out forwards;
        }
        
        @keyframes vanish {
            0% {
                opacity: 1;
                transform: translateX(0);
            }
            100% {
                opacity: 0;
                transform: translateX(-100%);
                height: 0;
                margin: 0;
                padding: 0;
            }
        }
        
        .loading {
            text-align: center;
            padding: 60px 20px;
            color: #ff6600;
            font-size: 1.1em;
            background: #1a1a1a;
            border-radius: 8px;
            border: 1px solid #333333;
        }
        
        .error {
            background: #1a1a1a;
            color: #ff3333;
            padding: 20px;
            border-radius: 8px;
            margin: 20px 0;
            border: 1px solid #ff3333;
            font-weight: 500;
        }
        
        @media (max-width: 768px) {
            .container {
                padding: 15px;
            }
            
            .logo {
                font-size: 2em;
            }
            
            .title {
                font-size: 1.5em;
            }
            
            .stats {
                flex-direction: column;
                gap: 20px;
                padding: 15px;
            }
            
            .controls {
                flex-direction: column;
                align-items: stretch;
                gap: 10px;
            }
            
            .contract-details {
                grid-template-columns: 1fr;
                gap: 15px;
            }
            
            .contract-header {
                flex-direction: column;
                gap: 10px;
                align-items: flex-start;
            }
            
            .contract-body {
                padding: 20px;
            }
        }
        
        .contract-link {
            display: inline-block;
            background: linear-gradient(135deg, #ff6600, #ff8533);
            color: #000000;
            text-decoration: none;
            padding: 6px 12px;
            border-radius: 4px;
            font-size: 0.85em;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            transition: all 0.3s ease;
            border: 1px solid #ff6600;
            box-shadow: 0 2px 4px rgba(255, 102, 0, 0.2);
        }
        
        .contract-link:hover {
            background: linear-gradient(135deg, #ff8533, #ff6600);
            transform: translateY(-1px);
            box-shadow: 0 4px 8px rgba(255, 102, 0, 0.3);
            color: #000000;
        }
        
        .contract-link:active {
            transform: translateY(0);
            box-shadow: 0 2px 4px rgba(255, 102, 0, 0.2);
        }
        
        .document-buttons {
            display: flex;
            gap: 8px;
            flex-wrap: wrap;
        }
        
        .document-link {
            display: inline-block;
            text-decoration: none;
            padding: 4px 8px;
            border-radius: 3px;
            font-size: 0.75em;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.3px;
            transition: all 0.3s ease;
            border: 1px solid;
        }
        
        .document-link.pliego {
            background: linear-gradient(135deg, #4CAF50, #66BB6A);
            color: #000000;
            border-color: #4CAF50;
            box-shadow: 0 1px 3px rgba(76, 175, 80, 0.3);
        }
        
        .document-link.pliego:hover {
            background: linear-gradient(135deg, #66BB6A, #4CAF50);
            transform: translateY(-1px);
            box-shadow: 0 2px 6px rgba(76, 175, 80, 0.4);
            color: #000000;
        }
        
        .document-link.anuncio {
            background: linear-gradient(135deg, #2196F3, #42A5F5);
            color: #000000;
            border-color: #2196F3;
            box-shadow: 0 1px 3px rgba(33, 150, 243, 0.3);
        }
        
        .document-link.anuncio:hover {
            background: linear-gradient(135deg, #42A5F5, #2196F3);
            transform: translateY(-1px);
            box-shadow: 0 2px 6px rgba(33, 150, 243, 0.4);
            color: #000000;
        }
        
        .document-link:active {
            transform: translateY(0);
        }
        
        .no-docs {
            color: #888888;
            font-style: italic;
            font-size: 0.85em;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">
                <span class="logo-text">Dashboard</span>
            </div>
            <div class="title">Contratos del Sector Público</div>
        </div>
        
        <div class="stats">
            <div class="stat">
                <div class="stat-number" id="totalContracts">-</div>
                <div class="stat-label">Total Contracts</div>
            </div>
            <div class="stat">
                <div class="stat-number" id="newContracts">-</div>
                <div class="stat-label">New Today</div>
            </div>
        </div>
        
        <div class="controls">
            <input type="text" class="search" id="searchInput" placeholder="Search contracts...">
            <button class="btn btn-primary" onclick="refreshData()">Refresh</button>
            <a href="/history" class="btn btn-primary">View History</a>
            <button class="btn btn-danger" onclick="deleteAll()">Delete All</button>
        </div>
        
        <div class="status-changes" id="statusChangesContainer" style="display: none;">
            <h3 style="color: #ff6600; margin-bottom: 15px;">Recent Status Changes</h3>
            <div id="statusChangesList"></div>
        </div>
        
        <div class="contracts" id="contractsContainer">
            <div class="loading">Loading contracts...</div>
        </div>
    </div>

    <script>
        let contracts = [];
        
        function loadContracts() {
            fetch('/api/contracts')
                .then(response => response.json())
                .then(data => {
                    contracts = data;
                    displayContracts(contracts);
                    loadStats();
                    loadStatusChanges();
                })
                .catch(error => {
                    document.getElementById('contractsContainer').innerHTML = 
                        '<div class="error">Error loading contracts: ' + error.message + '</div>';
                });
        }
        
        function loadStats() {
            fetch('/api/stats')
                .then(response => response.json())
                .then(data => {
                    document.getElementById('totalContracts').textContent = data.total;
                    document.getElementById('newContracts').textContent = data.newToday;
                })
                .catch(error => console.error('Error loading stats:', error));
        }
        
        function loadStatusChanges() {
            fetch('/api/status-changes')
                .then(response => response.json())
                .then(data => {
                    displayStatusChanges(data);
                })
                .catch(error => console.error('Error loading status changes:', error));
        }
        
        function displayStatusChanges(statusChanges) {
            const container = document.getElementById('statusChangesContainer');
            const list = document.getElementById('statusChangesList');
            
            if (statusChanges.length === 0) {
                container.style.display = 'none';
                return;
            }
            
            container.style.display = 'block';
            
            // Get dismissed changes from localStorage
            const dismissedChanges = JSON.parse(localStorage.getItem('dismissedStatusChanges') || '[]');
            
            // Filter out dismissed changes
            const visibleChanges = statusChanges.filter(change => !dismissedChanges.includes(change.id));
            
            if (visibleChanges.length === 0) {
                container.style.display = 'none';
                return;
            }
            
            list.innerHTML = visibleChanges.map((change, index) => {
                return '<div class="status-change-item" data-change-id="' + change.id + '">' +
                    '<div class="status-change-info">' +
                        '<div class="status-change-contract">' + change.contract_id + '</div>' +
                        '<div class="status-change-details">' +
                            '<span>' + change.old_status + '</span>' +
                            '<span class="status-change-arrow">→</span>' +
                            '<span>' + change.new_status + '</span>' +
                        '</div>' +
                    '</div>' +
                    '<div class="status-change-time">' + new Date(change.changed_at).toLocaleString() + '</div>' +
                    '<button class="status-change-checkmark" onclick="dismissChange(' + change.id + ')">✓</button>' +
                '</div>';
            }).join('');
        }
        
        function dismissChange(changeId) {
            const item = document.querySelector('[data-change-id="' + changeId + '"]');
            if (item) {
                // Add vanishing animation
                item.classList.add('vanishing');
                
                // Store in localStorage to persist the dismissed state
                const dismissedChanges = JSON.parse(localStorage.getItem('dismissedStatusChanges') || '[]');
                if (!dismissedChanges.includes(changeId)) {
                    dismissedChanges.push(changeId);
                    localStorage.setItem('dismissedStatusChanges', JSON.stringify(dismissedChanges));
                }
                
                // Remove the element after animation completes
                setTimeout(() => {
                    item.remove();
                    
                    // Check if there are any remaining status changes
                    const remainingItems = document.querySelectorAll('.status-change-item');
                    if (remainingItems.length === 0) {
                        document.getElementById('statusChangesContainer').style.display = 'none';
                    }
                }, 500);
            }
        }
        
        function getStatusClass(status) {
            const statusMap = {
                'publicada': 'publicada',
                'adjudicada': 'adjudicada',
                'anulada': 'anulada',
                'evaluación previa': 'evaluación-previa',
                'evaluacion previa': 'evaluación-previa',
                'resuelta': 'resuelta'
            };
            return statusMap[status.toLowerCase()] || status.toLowerCase().replace(/\s+/g, '-');
        }
        
        function displayContracts(contractsToShow) {
            const container = document.getElementById('contractsContainer');
            
            if (contractsToShow.length === 0) {
                container.innerHTML = '<div class="loading">No contracts found</div>';
                return;
            }
            
            container.innerHTML = contractsToShow.map(contract => 
            '<div class="contract">' +
                '<div class="contract-header">' +
                    '<div class="contract-id">' + contract.id + '</div>' +
                    '<div class="contract-actions">' +
                        '<div class="contract-status status-' + getStatusClass(contract.status) + '">' + contract.status + '</div>' +
                        '<button class="delete-contract-btn" onclick="deleteContract(\'' + contract.id + '\')" title="Delete contract">×</button>' +
                    '</div>' +
                '</div>' +
                '<div class="contract-body">' +
                    '<div class="contract-description">' + contract.description + '</div>' +
                    '<div class="contract-details">' +
                        '<div class="detail-item">' +
                            '<div class="detail-label">Type</div>' +
                            '<div>' + contract.contract_type + '</div>' +
                        '</div>' +
                        '<div class="detail-item">' +
                            '<div class="detail-label">Amount</div>' +
                            '<div class="amount">' + contract.amount + '</div>' +
                        '</div>' +
                        '<div class="detail-item">' +
                            '<div class="detail-label">Submission Date</div>' +
                            '<div>' + contract.submission_date + '</div>' +
                        '</div>' +
                        '<div class="detail-item">' +
                            '<div class="detail-label">Contracting Body</div>' +
                            '<div>' + contract.contracting_body + '</div>' +
                        '</div>' +
                        '<div class="detail-item">' +
                            '<div class="detail-label">Scraped At</div>' +
                            '<div>' + new Date(contract.scraped_at).toLocaleString() + '</div>' +
                        '</div>' +
                        '<div class="detail-item">' +
                            '<div class="detail-label">Documents</div>' +
                            '<div class="document-buttons">' +
                                (contract.pliego_link ? '<a href="' + contract.pliego_link + '" target="_blank" class="document-link pliego">Pliego</a>' : '') +
                                (contract.anuncio_link ? '<a href="' + contract.anuncio_link + '" target="_blank" class="document-link anuncio">Anuncio</a>' : '') +
                                (!contract.pliego_link && !contract.anuncio_link ? '<span class="no-docs">No disponible</span>' : '') +
                            '</div>' +
                        '</div>' +
                    '</div>' +
                '</div>' +
            '</div>'
        ).join('');
        }
        
        function refreshData() {
            loadContracts();
        }
        
        function deleteContract(contractId) {
            if (confirm('Are you sure you want to delete contract "' + contractId + '"? This action cannot be undone.')) {
                fetch('/api/delete-contract', { 
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ id: contractId })
                })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        loadContracts();
                    } else {
                        alert('Error deleting contract: ' + data.error);
                    }
                })
                .catch(error => {
                    alert('Error deleting contract: ' + error.message);
                });
            }
        }
        
        function deleteAll() {
            if (confirm('Are you sure you want to delete all contracts? This action cannot be undone.')) {
                fetch('/api/delete-all', { method: 'POST' })
                    .then(response => response.json())
                    .then(data => {
                        if (data.success) {
                            loadContracts();
                        } else {
                            alert('Error deleting contracts: ' + data.error);
                        }
                    })
                    .catch(error => {
                        alert('Error deleting contracts: ' + error.message);
                    });
            }
        }
        
        // Search functionality
        document.getElementById('searchInput').addEventListener('input', function(e) {
            const searchTerm = e.target.value.toLowerCase();
            const filtered = contracts.filter(contract => 
                contract.description.toLowerCase().includes(searchTerm) ||
                contract.id.toLowerCase().includes(searchTerm) ||
                contract.contracting_body.toLowerCase().includes(searchTerm)
            );
            displayContracts(filtered);
        });
        
        // Load data on page load
        loadContracts();
        
        // Auto-refresh every 30 seconds
        setInterval(loadStats, 30000);
    </script>
</body>
</html>`

	HistoryTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Historial de Cambios</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: #000000;
            color: #ffffff;
            line-height: 1.6;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }
        
        .header {
            text-align: center;
            margin-bottom: 40px;
            padding: 20px;
            background: #1a1a1a;
            border-radius: 8px;
            border: 1px solid #333333;
        }
        
        .logo {
            font-size: 2.5em;
            font-weight: bold;
            color: #ff6600;
            margin-bottom: 10px;
        }
        
        .title {
            font-size: 1.8em;
            color: #ffffff;
            margin-bottom: 10px;
        }
        
        .subtitle {
            color: #666666;
            font-size: 1em;
        }
        
        .back-button {
            display: inline-block;
            background: linear-gradient(135deg, #ff6600, #ff8533);
            color: #000000;
            text-decoration: none;
            padding: 10px 20px;
            border-radius: 6px;
            font-weight: 600;
            margin-bottom: 20px;
            transition: all 0.3s ease;
            border: 1px solid #ff6600;
        }
        
        .back-button:hover {
            background: linear-gradient(135deg, #ff8533, #ff6600);
            transform: translateY(-2px);
            box-shadow: 0 4px 8px rgba(255, 102, 0, 0.3);
        }
        
        .status-changes {
            background: #1a1a1a;
            border-radius: 8px;
            border: 1px solid #333333;
            padding: 20px;
        }
        
        .status-change-item {
            background: #000000;
            border-radius: 6px;
            padding: 15px;
            margin-bottom: 10px;
            border: 1px solid #333333;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        .status-change-info {
            flex: 1;
        }
        
        .status-change-contract {
            color: #ff6600;
            font-weight: bold;
            font-size: 1.1em;
            margin-bottom: 5px;
        }
        
        .status-change-details {
            color: #ffffff;
            font-size: 0.9em;
        }
        
        .status-change-arrow {
            color: #ff6600;
            margin: 0 10px;
        }
        
        .status-change-time {
            color: #666666;
            font-size: 0.8em;
            text-align: right;
        }
        
        .no-changes {
            text-align: center;
            padding: 60px 20px;
            color: #666666;
            font-size: 1.1em;
        }
    </style>
</head>
<body>
    <div class="container">
        <a href="/" class="back-button">← Back to Dashboard</a>
        
        <div class="header">
            <div class="title">Historial de Cambios</div>
        </div>
        
        <div class="status-changes">
            <div id="statusChangesList">
                {{if .StatusChanges}}
                    {{range .StatusChanges}}
                    <div class="status-change-item">
                        <div class="status-change-info">
                            <div class="status-change-contract">{{.ContractID}}</div>
                            <div class="status-change-details">
                                <span>{{.OldStatus}}</span>
                                <span class="status-change-arrow">→</span>
                                <span>{{.NewStatus}}</span>
                            </div>
                        </div>
                        <div class="status-change-time">{{.ChangedAt}}</div>
                    </div>
                    {{end}}
                {{else}}
                    <div class="no-changes">No status changes found</div>
                {{end}}
            </div>
        </div>
    </div>
</body>
</html>`
) 