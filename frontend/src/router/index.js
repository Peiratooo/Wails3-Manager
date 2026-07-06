import { createRouter, createWebHashHistory } from 'vue-router'


const routes = [
	{
		path: '/',
		name: 'home',
		component: () => import('../views/HomeView.vue'),
	},
    {
		path: '/project',
		name: 'project',
		component: () => import('../views/ProjectView.vue'),
	},
]

const router = createRouter({
	history: createWebHashHistory(),
	routes,
})

export default router
