document.addEventListener("DOMContentLoaded", function () {
    // State management
    const state = {
        isConsumerRunning: false,
        isDarkMode: false,
        isLoading: false,
        loadingMessage: "",
        notification: {
            message: "",
            type: "", // success, error
            visible: false,
        },
    };

    // DOM Elements
    const messageArea = document.getElementById("messageArea");
    const notificationMessage = document.getElementById("notificationMessage");
    const themeToggle = document.getElementById("themeToggle");
    const startButton = document.getElementById("startButton");
    const stopButton = document.getElementById("stopButton");
    const updateButton = document.getElementById("updateButton");
    const loadingArea = document.getElementById("loadingArea");
    const loadingMessageElement = document.getElementById("loadingMessage");
    const logsOutput = document.getElementById("logsContent");
    const body = document.body;
    const brokersInput = document.getElementById("brokers");
    const groupIdInput = document.getElementById("groupId");
    const topicInput = document.getElementById("topic");
    const minBytesInput = document.getElementById("minBytes");
    const maxBytesInput = document.getElementById("maxBytes");
    const messageLimitInput = document.getElementById("messageLimit");
    const postUrlInput = document.getElementById("postUrl");
    const portAppDisplay = document.getElementById("portAppDisplay");

    let messageTimeout;
    let retryCount = 0;
    const maxRetries = 5;

    // Load configuration on page load
    async function loadConfig() {
        try {
            const response = await fetch("/api/config");
            if (response.ok) {
                const config = await response.json();

                // Fill fields with config data
                brokersInput.value = config.brokers.join(", ");
                groupIdInput.value = config.group_id;
                topicInput.value = config.topic;
                minBytesInput.value = config.min_bytes;
                maxBytesInput.value = config.max_bytes;
                messageLimitInput.value = config.message_limit;
                postUrlInput.value = config.post_url;
                portAppDisplay.textContent = config.portApp;
            } else {
                console.error("Не удалось загрузить конфигурацию");
                showNotification("Ошибка загрузки конфигурации", "error");
            }
        } catch (error) {
            console.error("Ошибка при загрузке конфигурации:", error);
            showNotification("Ошибка при загрузке конфигурации", "error");
        }
    }

    // Show notification
    function showNotification(message, type = "success") {
        state.notification.message = message;
        state.notification.type = type;
        state.notification.visible = true;

        notificationMessage.textContent = message;
        messageArea.classList.add("visible");

        if (type === "error") {
            messageArea.classList.add("error");
        } else {
            messageArea.classList.remove("error");
        }

        clearTimeout(messageTimeout);
        messageTimeout = setTimeout(() => {
            messageArea.classList.remove("visible");
            state.notification.visible = false;
        }, 3000);
    }

    // Start consumer
    async function startConsumer() {
        state.isLoading = true;
        state.loadingMessage = "Запускаем консьюмер...";
        updateUI();
        try {
            const response = await fetch("/api/start", { method: "POST" });
            const message = await response.text();
            if (response.ok) {
                state.isConsumerRunning = true;
                showNotification(message, "success");
            } else {
                showNotification(message, "error");
            }
        } catch (error) {
            showNotification(`Ошибка при запуске: ${error.message}`, "error");
        } finally {
            state.isLoading = false;
            state.loadingMessage = "";
            updateUI();
        }
    }

    // Stop consumer
    async function stopConsumer() {
        state.isLoading = true;
        state.loadingMessage = "Останавливаем консьюмер...";
        updateUI();
        try {
            const response = await fetch("/api/stop", { method: "POST" });
            const message = await response.text();
            if (response.ok) {
                state.isConsumerRunning = false;
                showNotification(message, "success");
            } else {
                showNotification(message, "error");
            }
        } catch (error) {
            showNotification(`Ошибка при остановке: ${error.message}`, "error");
        } finally {
            state.isLoading = false;
            state.loadingMessage = "";
            updateUI();
        }
    }

    // Update configuration
    async function updateConfig() {
        state.isLoading = true;
        state.loadingMessage = "Обновляем конфигурацию...";
        updateUI();

        const config = {
            brokers: brokersInput.value.split(",").map((b) => b.trim()),
            group_id: groupIdInput.value,
            topic: topicInput.value,
            min_bytes: parseInt(minBytesInput.value) || 0,
            max_bytes: parseInt(maxBytesInput.value) || 0,
            message_limit: parseInt(messageLimitInput.value) || 0,
            post_url: postUrlInput.value,
            portApp: portAppDisplay.textContent,
        };

        try {
            const response = await fetch("/api/config/update", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(config),
            });
            const message = await response.text();
            showNotification(message, response.ok ? "success" : "error");
            if (response.ok) {
                addLogEntry("Configuration updated at " + new Date().toLocaleTimeString());
            }
        } catch (error) {
            showNotification(`Ошибка при обновлении: ${error.message}`, "error");
        } finally {
            state.isLoading = false;
            state.loadingMessage = "";
            updateUI();
        }
    }

    // Update UI
    function updateUI() {
        loadingArea.classList.toggle("visible", state.isLoading);
        loadingMessageElement.textContent = state.loadingMessage;

        if (state.isConsumerRunning) {
            startButton.style.display = "none";
            stopButton.style.display = "flex";
        } else {
            startButton.style.display = "flex";
            stopButton.style.display = "none";
        }

        if (state.isDarkMode) {
            body.classList.add("dark-mode");
        } else {
            body.classList.remove("dark-mode");
        }
    }

    // Add log entry
    function addLogEntry(message) {
        const logEntry = document.createElement("div");
        logEntry.textContent = `[${new Date().toLocaleTimeString()}] ${message}`;
        logEntry.style.borderBottom = "1px solid #e0e0e0";
        logEntry.style.padding = "8px 0";

        if (state.isDarkMode) {
            logEntry.style.borderBottom = "1px solid #3A3F55";
        }

        logsOutput.prepend(logEntry);

        if (logsOutput.children.length > 100) {
            logsOutput.removeChild(logsOutput.lastChild);
        }
    }

    // Setup WebSocket
    function setupWebSocket() {
        const wsUrl = `ws://${window.location.hostname}:8082/ws`;
        const socket = new WebSocket(wsUrl);

        socket.onopen = () => {
            addLogEntry(`[System] WebSocket подключен к ${wsUrl}`);
        };

        socket.onmessage = (event) => {
            const message = event.data;

            if (message.includes("Consumer stopped")) {
                state.isConsumerRunning = false;
                updateUI();
            } else if (message.includes("Consumer started")) {
                state.isConsumerRunning = true;
                updateUI();
            }

            addLogEntry(message);
        };

        socket.onerror = () => {
            addLogEntry("[System] Ошибка WebSocket соединения.");
        };

        socket.onclose = () => {
            addLogEntry("[System] WebSocket соединение закрыто.");
        };
    }

    // Toggle theme
    function toggleTheme() {
        state.isDarkMode = !state.isDarkMode;
        updateTheme();
    }

    function updateTheme() {
        if (state.isDarkMode) {
            body.classList.add("dark-mode");
        } else {
            body.classList.remove("dark-mode");
        }
    }

    // Event Listeners
    themeToggle.addEventListener("click", toggleTheme);
    startButton.addEventListener("click", startConsumer);
    stopButton.addEventListener("click", stopConsumer);
    updateButton.addEventListener("click", (e) => {
        e.preventDefault();
        updateConfig();
    });

    // Initialize
    updateUI();
    setupWebSocket();
    loadConfig();
    addLogEntry("Dashboard initialized");
});