import { createRouter, createWebHistory } from 'vue-router';
import ChatRoom from '../views/ChatRoom.vue';

const routes = [
  { path: '/', component: ChatRoom },
  { path: '/chat', component: ChatRoom }
];

const router = createRouter({
  history: createWebHistory(),
  routes
});

export default router;