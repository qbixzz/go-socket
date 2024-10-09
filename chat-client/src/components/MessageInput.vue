<template>
    <div class="message-input">
      <input v-model="message" @keyup.enter="sendMessage" placeholder="Type a message" style="height: 50px;" />
      <button @click="sendMessage" style="height: 50px;">
        <i class="fas fa-paper-plane"></i>
      </button>
    </div>
  </template>
  
  <script>
  export default {
    data() {
      return {
        message: ''
      };
    },
    methods: {
      sendMessage() {
        if (this.message.trim()) {
          this.$emit('send-message', this.message);
          this.message = '';
        }
      }
    }
  };
  </script>
  
  <style scoped>
  .message-input {
    padding: 10px;
    border-top: 1px solid #ccc;
    display: flex;
    align-items: center;
  }
  input {
    flex: 1;
    padding: 10px;
    box-sizing: border-box;
    margin-right: 10px;
  }
  button {
    padding: 10px;
    box-sizing: border-box;
  }
  </style>