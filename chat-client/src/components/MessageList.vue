<template>
    <div class="message-list">
      <div
        v-for="(message, index) in messages"
        :key="index"
        :class="['message', message.event === 'client-message' ? 'message-right' : 'message-left']"
      >
        {{ message.text }}
      </div>
    </div>
  </template>
  
  <script>
  export default {
    props: {
      messages: {
        type: Array,
        required: true
      },
      clientId: {
        type: String,
        required: true
      }
    }
  };
  </script>
  
  <style scoped>
  .message-list {
    display: flex;
    flex-direction: column;
  }
  
  .message {
    max-width: 60%;
    font-size: 2em;
    padding: 5px 15px;
    margin: 5px 0;
    border-radius: 10px;
    word-wrap: break-word;
  }
  
  .message-right {
    align-self: flex-end;
    background-color: #dcf8c6;
    color: #000;
  }
  
  .message-left {
    align-self: flex-start;
    background-color: #fff;
    color: #000;
    border: 1px solid #ddd;
  }
  </style>