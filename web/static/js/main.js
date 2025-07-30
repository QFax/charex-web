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
        name.textContent = card.data.name;
        cardDiv.appendChild(name);

        if (card.data.creator) {
            const creator = document.createElement('p');
            creator.textContent = `By: ${card.data.creator}`;
            cardDiv.appendChild(creator);
        }
        
        const description = document.createElement('p');
        description.textContent = card.data.description;
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
        console.log("Initializing application");
        const data = await fetchCards();
        if (data && data.sources) {
            allCards = data.sources.flatMap(s => s.cards.map(c => ({...c, data: c.data, source: s.name })));
            renderCards();
            console.log("Cards loaded, connecting to WebSocket.");
        } else {
            console.warn("Failed to load initial card data.");
            // Optionally, display a user-friendly message here
        }
        window.ws.connect();
        
        window.ws.on('status', (payload) => {
            console.log('Status update:', payload);
            // You could display these status messages in a dedicated area if you wish.
            if (payload.status === 'error') {
                errorMessage.textContent = payload.message;
            } else {
                errorMessage.textContent = payload.message; // Or a different element
            }
        });

        window.ws.on('new_card', (payload) => {
            errorMessage.textContent = ''; // Clear previous errors
            const newCard = {...payload.card, data: payload.card.data, source: payload.source};
            allCards.unshift(newCard); // Add to the beginning of the list
            renderCards();
        });

        window.ws.on('error', (payload) => {
            errorMessage.textContent = payload.message;
        });
    }

    urlForm.addEventListener('submit', (e) => {
        e.preventDefault();
        const url = urlInput.value.trim();
        if (url) {
            window.ws.sendURLForExtraction(url);
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