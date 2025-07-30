window.ws = {
    socket: null,
    _handlers: {},

    on: function(eventType, handler) {
        if (!this._handlers[eventType]) {
            this._handlers[eventType] = [];
        }
        this._handlers[eventType].push(handler);
    },

    off: function(eventType, handler) {
        if (this._handlers[eventType]) {
            this._handlers[eventType] = this._handlers[eventType].filter(h => h !== handler);
        }
    },

    _emit: function(eventType, data) {
        if (this._handlers[eventType]) {
            this._handlers[eventType].forEach(handler => handler(data));
        }
    },

    connect: function() {
        const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${wsProtocol}//${window.location.host}/ws`;

        this.socket = new WebSocket(wsUrl);

        this.socket.onopen = () => {
            console.log('WebSocket connection established');
            this._emit('open');
        };

        this.socket.onmessage = (event) => {
            const message = JSON.parse(event.data);
            this._emit(message.type, message.payload);
        };

        this.socket.onclose = () => {
            console.log('WebSocket connection closed. Reconnecting in 2 seconds...');
            this._emit('close');
            setTimeout(() => this.connect(), 2000);
        };

        this.socket.onerror = (error) => {
            console.error('WebSocket error:', error);
            this._emit('error', { message: 'WebSocket connection error.' });
        };
    },

    sendURLForExtraction: function(url) {
        if (this.socket && this.socket.readyState === WebSocket.OPEN) {
            const message = {
                type: 'extract_sakura',
                payload: { url }
            };
            this.socket.send(JSON.stringify(message));
        } else {
            console.error('WebSocket is not connected.');
            this._emit('error', { message: 'Cannot send message. WebSocket is not connected.' });
        }
    }
};