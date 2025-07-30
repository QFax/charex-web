let socket;

function connectWebSocket(onMessageCallback) {
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${wsProtocol}//${window.location.host}/ws`;

    socket = new WebSocket(wsUrl);

    socket.onopen = () => {
        console.log('WebSocket connection established');
    };

    socket.onmessage = (event) => {
        const message = JSON.parse(event.data);
        onMessageCallback(message);
    };

    socket.onclose = () => {
        console.log('WebSocket connection closed. Reconnecting in 2 seconds...');
        setTimeout(() => connectWebSocket(onMessageCallback), 2000);
    };

    socket.onerror = (error) => {
        console.error('WebSocket error:', error);
        const errorMessage = document.getElementById('error-message');
        errorMessage.textContent = 'WebSocket connection error.';
    };
}

function sendURLForExtraction(url) {
    if (socket && socket.readyState === WebSocket.OPEN) {
        const message = {
            type: 'extract_sakura',
            payload: { url }
        };
        socket.send(JSON.stringify(message));
    } else {
        console.error('WebSocket is not connected.');
        const errorMessage = document.getElementById('error-message');
        errorMessage.textContent = 'Cannot send message. WebSocket is not connected.';
    }
}