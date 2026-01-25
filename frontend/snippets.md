# JS code on the frontend to listen for events on the SSE
Should be sufficient with cookies:
```
const eventSource = new EventSource("/api/visitors/events");

eventSource.onmessage = (event) => {
    if (event.data === "refresh_queue") {
        console.log("New visitor joined! Fetching updated list...");
        fetchQueueList(); // Call your existing GET /api/visitors/queue
    }
};
```
# Longer listener snippet without cookies
```
function setupQueueListener() {
    // 1. Initialize the connection
    // Note: If using token auth via URL, use: `/api/visitors/events?token=${myToken}`
    const eventSource = new EventSource("/api/visitors/events");

    // 2. Listen for the 'refresh_queue' message from Go
    eventSource.onmessage = (event) => {
        if (event.data === "refresh_queue") {
            console.log("Queue update received from server...");
            fetchQueueData(); // Your function that calls GET /api/visitors/queue
        }
    };

    // 3. Handle connection errors (e.g., server restart)
    eventSource.onerror = (err) => {
        console.error("EventSource failed:", err);
        // The browser automatically tries to reconnect, 
        // but you might want to log it for debugging.
    };
    
    // Optional: Close connection when leaving the page
    window.addEventListener('beforeunload', () => {
        eventSource.close();
    });
}

/**
 * Example function to update your UI
 */
async function fetchQueueData() {
    try {
        const response = await fetch("/api/visitors/queue", {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('access_token')}`
            }
        });
        const data = await response.json();
        
        // Update your DOM/React/Vue state here
        renderQueueTable(data);
    } catch (error) {
        console.error("Failed to fetch queue:", error);
    }
}

// Start listening when the dashboard loads
setupQueueListener();
```