async function fetchCards() {
    try {
        const response = await fetch('/api/cards');
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        return await response.json();
    } catch (error) {
        console.error('Failed to fetch cards:', error);
        const errorMessage = document.getElementById('error-message');
        errorMessage.textContent = 'Could not load character cards. Is the server running?';
        return null;
    }
}