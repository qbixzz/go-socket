<template>
  <div class="chat-window">
    <div class="header">
      <h1>Go Socket Chat</h1>
      <p class="clock">{{ serverTime }}</p>
    </div>
    <div class="message-list-container">
      <MessageList :messages="messages" :clientId="clientId" />
    </div>
    <div class="message-input-container">
      <MessageInput @send-message="sendMessage" />
    </div>
  </div>
</template>
  
  <script>
  import MessageList from './MessageList.vue';
  import MessageInput from './MessageInput.vue';
  
  export default {
    components: {
      MessageList,
      MessageInput
    },
    data() {
      return {
        messages: [],
        clientId: 'client-' + Math.random().toString(36).substr(2, 9),
        websocket: null,
        serverTime: '',
      };
    },
    methods: {
    connectWebSocket() {
      // const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsUrl = `wss://events.controldata.co.th/chat-server/ws/?clientId=${this.clientId}`;
      this.websocket = new WebSocket(wsUrl);

      this.websocket.onopen = () => {
        console.log('WebSocket is open now.');
      };

      this.websocket.onclose = () => {
        console.log('WebSocket is closed now.');
      };

      this.websocket.onerror = (error) => {
        console.error('WebSocket error observed:', error);
      };

      this.websocket.onmessage = this.receiveMessage;
    },
    connectSSE() {
      const eventSource = new EventSource('http://localhost:8081/sse');
      eventSource.onmessage = (event) => {
        this.serverTime = event.data;
      };
      eventSource.onerror = (error) => {
        console.error('EventSource failed:', error);
        eventSource.close();
      };
    },
    sendMessage(message) {
      if (message === 'connect') {
        this.connectWebSocket();
        return;
      }
      if (message === 'disconnect') {
        this.websocket.close();
        return;
      }
      if (message === 'clear') {
        this.messages = [];
        return;
      }
      if (message === 'connectsse') {
        this.connectSSE();
        return;
      }
      if (this.websocket.readyState === WebSocket.OPEN) {
        const msg = { text: message, sender: this.clientId, event: 'client-message' }
        this.messages.push(msg);
        this.websocket.send(JSON.stringify(msg));
      } else {
        console.error('WebSocket is not open. Ready state:', this.websocket.readyState);
      }
    },
    receiveMessage(event) {
        const message = JSON.parse(event.data);
        if (message.event === 'server-message' && message.sender !== this.clientId) {
          this.messages.push(message);
          console.log('Received message:', this.messages);
        }
    }
  },
  mounted() {
    this.connectWebSocket();
    this.connectSSE();
  },
  };
  </script>
  
  <style scoped>
  .header {
  padding: 10px;
  background-color: #4CAF50;
  color: white;
  text-align: left;
  display: flex;
  justify-content: space-between;
  align-items: center;
  }

  .clock {
  font-size: 1em;
  color: #333;
  }

  .chat-window {
    display: flex;
    flex-direction: column;
    height: 100vh;
    background-color: #f5f5f5;
    border: 1px solid #ddd;
    overflow: hidden;
  }
  
  .message-list-container {
    flex: 1;
    overflow-y: auto;
    padding: 10px;
    background-color: #fff;
  }
  
  .message-input-container {
    padding: 10px;
    background-color: #f1f1f1;
    border-top: 1px solid #ddd;
  }
  
  .message-list-container::-webkit-scrollbar {
    width: 8px;
  }
  
  .message-list-container::-webkit-scrollbar-thumb {
    background-color: #ccc;
    border-radius: 4px;
  }
  
  .message-list-container::-webkit-scrollbar-track {
    background-color: #f5f5f5;
  }
  </style>