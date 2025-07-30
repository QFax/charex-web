document.addEventListener('DOMContentLoaded', () => {
    let allCards = [];
    let currentSort = { key: 'create_date', direction: 'desc' };

    const cardContainer = document.getElementById('card-container');
    const urlForm = document.getElementById('url-form');
    const urlInput = document.getElementById('url-input');
    const errorMessage = document.getElementById('error-message');

    function createCardElement(card) {
        const cardDiv = document.createElement('div');
        cardDiv.className = 'card';

        const name = document.createElement('h3');
        name.textContent = card.name;
        cardDiv.appendChild(name);

        if (card.creator) {
            const creator = document.createElement('p');
            creator.textContent = `By: ${card.creator}`;
            cardDiv.appendChild(creator);
        }
        
        const description = document.createElement('p');
        description.textContent = card.description;
        cardDiv.appendChild(description);

        return cardDiv;
    }

    function renderCards() {
        cardContainer.innerHTML = '';
        const groupedBySource = allCards.reduce((acc, card) => {
            const source = card.source || 'unknown';
            if (!acc[source]) {
                acc[source] = [];
            }
            acc[source].push(card);
            return acc;
        }, {});

        Object.keys(groupedBySource).sort().forEach(source => {
            const sourceSection = document.createElement('div');
            sourceSection.className = 'source-section';
            
            const sourceTitle = document.createElement('h2');
            sourceTitle.textContent = source;
            sourceSection.appendChild(sourceTitle);

            const sourceCardContainer = document.createElement('div');
            sourceCardContainer.className = 'card-grid';
            
            let cards = groupedBySource[source];

            // Sorting logic
            cards.sort((a, b) => {
                let valA = a[currentSort.key];
                let valB = b[currentSort.key];

                if (currentSort.key === 'create_date') {
                    valA = new Date(valA).getTime();
                    valB = new Date(valB).getTime();
                }

                if (valA < valB) return currentSort.direction === 'asc' ? -1 : 1;
                if (valA > valB) return currentSort.direction === 'asc' ? 1 : -1;
                return 0;
            });

            cards.forEach(card => {
                sourceCardContainer.appendChild(createCardElement(card));
            });

            sourceSection.appendChild(sourceCardContainer);
            cardContainer.appendChild(sourceSection);
        });
    }

    async function initialize() {
        const data = await fetchCards();
        if (data && data.sources) {
            allCards = data.sources.flatMap(s => s.cards.map(c => ({...c, source: s.name })));
            renderCards();
        }
        connectWebSocket(handleWebSocketMessage);
    }

    function handleWebSocketMessage(message) {
        errorMessage.textContent = ''; // Clear previous errors
        if (message.type === 'new_card' && message.payload) {
            const newCard = {...message.payload.card, source: message.payload.source};
            allCards.push(newCard);
            renderCards();
        } else if (message.type === 'error' && message.payload) {
            errorMessage.textContent = message.payload.message;
        }
    }

    urlForm.addEventListener('submit', (e) => {
        e.preventDefault();
        const url = urlInput.value.trim();
        if (url) {
            sendURLForExtraction(url);
            urlInput.value = '';
        }
    });

    function setSort(key, direction) {
        currentSort = { key, direction };
        renderCards();
    }

    document.getElementById('sort-name-asc').addEventListener('click', () => setSort('name', 'asc'));
    document.getElementById('sort-name-desc').addEventListener('click', () => setSort('name', 'desc'));
    document.getElementById('sort-date-asc').addEventListener('click', () => setSort('create_date', 'asc'));
    document.getElementById('sort-date-desc').addEventListener('click', () => setSort('create_date', 'desc'));

    initialize();
});