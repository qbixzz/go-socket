<template>
    <div class="message-input">
      <input v-model="message" @keyup.enter="sendMessage" placeholder="Type a message" />
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
  }
  input {
    width: 100%;
    padding: 10px;
    box-sizing: border-box;
  }
  </style>