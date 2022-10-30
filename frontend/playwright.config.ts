import type { PlaywrightTestConfig } from '@playwright/test'

const encode = (obj: any) => {
	return Buffer.from(JSON.stringify(obj)).toString('base64')
}
const config: PlaywrightTestConfig = {
	webServer: {
		command: 'npm run build && npm run preview',
		port: 4173
	},
	testDir: './tests/',
	use: {
		baseURL: 'http://localhost:8080',
		extraHTTPHeaders: {
			DEV_GAME_OPTIONS: encode({ seed: 123, state: 123 }),
			DEV_USERNAME: 'PlayWright Testing',
			Authorization: ('pw-test-' + Math.random() * 1e6).slice(0, 21)
		}
	}
}

export default config
