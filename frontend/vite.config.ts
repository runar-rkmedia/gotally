import { sveltekit } from '@sveltejs/kit/vite'
import type { UserConfig } from 'vite'
import { networkInterfaces } from 'os'

if (!process.env.VITE_DEV_API) {
	const ip = getIP()
	if (ip) {
		process.env.VITE_DEV_API = 'http://' + ip + ':8080'
	}
}

console.log('ips', getIP())

const config: UserConfig = {
	plugins: [sveltekit()]
}

function getIP() {
	const nets = networkInterfaces()

	if (!nets) {
		return ''
	}

	for (const name of Object.keys(nets)) {
		const n = nets[name]
		if (!n) {
			continue
		}
		for (const net of n) {
			// Skip over non-IPv4 and internal (i.e. 127.0.0.1) addresses
			// 'IPv4' is in Node <= 17, from 18 it's a number 4 or 6
			const familyV4Value = typeof net.family === 'string' ? 'IPv4' : 4
			if (net.family === familyV4Value && !net.internal) {
				return net.address
			}
		}
	}
}

export default config
