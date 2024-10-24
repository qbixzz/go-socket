<template>
  <div class="chat-window">
    <div class="header">
      <h1>Go Socket Chat</h1>
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
    sendMessage(message) {
      if (message === 'connect') {
        this.connectWebSocket();
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
  },
  };
  </script>
  
  <style scoped>
  .header {
  padding: 10px;
  background-color: #4CAF50;
  color: white;
  text-align: left;
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