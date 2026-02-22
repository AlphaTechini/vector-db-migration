
// this file is generated — do not edit it


/// <reference types="@sveltejs/kit" />

/**
 * Environment variables [loaded by Vite](https://vitejs.dev/guide/env-and-mode.html#env-files) from `.env` files and `process.env`. Like [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private), this module cannot be imported into client-side code. This module only includes variables that _do not_ begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) _and do_ start with [`config.kit.env.privatePrefix`](https://svelte.dev/docs/kit/configuration#env) (if configured).
 * 
 * _Unlike_ [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private), the values exported from this module are statically injected into your bundle at build time, enabling optimisations like dead code elimination.
 * 
 * ```ts
 * import { API_KEY } from '$env/static/private';
 * ```
 * 
 * Note that all environment variables referenced in your code should be declared (for example in an `.env` file), even if they don't have a value until the app is deployed:
 * 
 * ```
 * MY_FEATURE_FLAG=""
 * ```
 * 
 * You can override `.env` values from the command line like so:
 * 
 * ```sh
 * MY_FEATURE_FLAG="enabled" npm run dev
 * ```
 */
declare module '$env/static/private' {
	export const SELKIES_INTERPOSER: string;
	export const LANGUAGE: string;
	export const npm_config_user_agent: string;
	export const S6_STAGE2_HOOK: string;
	export const HOSTNAME: string;
	export const SSH_AGENT_PID: string;
	export const npm_node_execpath: string;
	export const SHLVL: string;
	export const npm_config_noproxy: string;
	export const HOME: string;
	export const OLDPWD: string;
	export const DISABLE_DRI3: string;
	export const npm_package_json: string;
	export const PANEL_GDK_CORE_DEVICE_EVENTS: string;
	export const HOMEBREW_PREFIX: string;
	export const CUSTOM_WS_PORT: string;
	export const S6_VERBOSITY: string;
	export const DISABLE_ZINK: string;
	export const NODE_NO_WARNINGS: string;
	export const npm_config_userconfig: string;
	export const npm_config_local_prefix: string;
	export const DBUS_SESSION_BUS_ADDRESS: string;
	export const COLORTERM: string;
	export const PGID: string;
	export const COLOR: string;
	export const LSIO_FIRST_PARTY: string;
	export const TITLE: string;
	export const INFOPATH: string;
	export const S6_CMD_WAIT_FOR_SERVICES_MAXTIME: string;
	export const WINDOWID: string;
	export const _: string;
	export const npm_config_prefix: string;
	export const npm_config_npm_version: string;
	export const NVIDIA_DRIVER_CAPABILITIES: string;
	export const TERM: string;
	export const npm_config_cache: string;
	export const OPENCLAW_NODE_OPTIONS_READY: string;
	export const npm_config_node_gyp: string;
	export const PATH: string;
	export const SESSION_MANAGER: string;
	export const PULSE_RUNTIME_PATH: string;
	export const HOMEBREW_CELLAR: string;
	export const NODE: string;
	export const npm_package_name: string;
	export const PERL5LIB: string;
	export const XDG_RUNTIME_DIR: string;
	export const DISPLAY: string;
	export const LD_PRELOAD: string;
	export const LANG: string;
	export const OPENCLAW_GATEWAY_PORT: string;
	export const npm_lifecycle_script: string;
	export const PUID: string;
	export const SSH_AUTH_SOCK: string;
	export const npm_package_version: string;
	export const npm_lifecycle_event: string;
	export const VIRTUAL_ENV: string;
	export const npm_config_globalconfig: string;
	export const npm_config_init_module: string;
	export const PWD: string;
	export const npm_execpath: string;
	export const npm_config_global_prefix: string;
	export const HOMEBREW_REPOSITORY: string;
	export const npm_command: string;
	export const TZ: string;
	export const VTE_VERSION: string;
	export const START_DOCKER: string;
	export const GOOGLE_API_KEY: string;
	export const OPENCLAW_PATH_BOOTSTRAPPED: string;
	export const INIT_CWD: string;
	export const EDITOR: string;
	export const NODE_ENV: string;
}

/**
 * Similar to [`$env/static/private`](https://svelte.dev/docs/kit/$env-static-private), except that it only includes environment variables that begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) (which defaults to `PUBLIC_`), and can therefore safely be exposed to client-side code.
 * 
 * Values are replaced statically at build time.
 * 
 * ```ts
 * import { PUBLIC_BASE_URL } from '$env/static/public';
 * ```
 */
declare module '$env/static/public' {
	
}

/**
 * This module provides access to runtime environment variables, as defined by the platform you're running on. For example if you're using [`adapter-node`](https://github.com/sveltejs/kit/tree/main/packages/adapter-node) (or running [`vite preview`](https://svelte.dev/docs/kit/cli)), this is equivalent to `process.env`. This module only includes variables that _do not_ begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) _and do_ start with [`config.kit.env.privatePrefix`](https://svelte.dev/docs/kit/configuration#env) (if configured).
 * 
 * This module cannot be imported into client-side code.
 * 
 * ```ts
 * import { env } from '$env/dynamic/private';
 * console.log(env.DEPLOYMENT_SPECIFIC_VARIABLE);
 * ```
 * 
 * > [!NOTE] In `dev`, `$env/dynamic` always includes environment variables from `.env`. In `prod`, this behavior will depend on your adapter.
 */
declare module '$env/dynamic/private' {
	export const env: {
		SELKIES_INTERPOSER: string;
		LANGUAGE: string;
		npm_config_user_agent: string;
		S6_STAGE2_HOOK: string;
		HOSTNAME: string;
		SSH_AGENT_PID: string;
		npm_node_execpath: string;
		SHLVL: string;
		npm_config_noproxy: string;
		HOME: string;
		OLDPWD: string;
		DISABLE_DRI3: string;
		npm_package_json: string;
		PANEL_GDK_CORE_DEVICE_EVENTS: string;
		HOMEBREW_PREFIX: string;
		CUSTOM_WS_PORT: string;
		S6_VERBOSITY: string;
		DISABLE_ZINK: string;
		NODE_NO_WARNINGS: string;
		npm_config_userconfig: string;
		npm_config_local_prefix: string;
		DBUS_SESSION_BUS_ADDRESS: string;
		COLORTERM: string;
		PGID: string;
		COLOR: string;
		LSIO_FIRST_PARTY: string;
		TITLE: string;
		INFOPATH: string;
		S6_CMD_WAIT_FOR_SERVICES_MAXTIME: string;
		WINDOWID: string;
		_: string;
		npm_config_prefix: string;
		npm_config_npm_version: string;
		NVIDIA_DRIVER_CAPABILITIES: string;
		TERM: string;
		npm_config_cache: string;
		OPENCLAW_NODE_OPTIONS_READY: string;
		npm_config_node_gyp: string;
		PATH: string;
		SESSION_MANAGER: string;
		PULSE_RUNTIME_PATH: string;
		HOMEBREW_CELLAR: string;
		NODE: string;
		npm_package_name: string;
		PERL5LIB: string;
		XDG_RUNTIME_DIR: string;
		DISPLAY: string;
		LD_PRELOAD: string;
		LANG: string;
		OPENCLAW_GATEWAY_PORT: string;
		npm_lifecycle_script: string;
		PUID: string;
		SSH_AUTH_SOCK: string;
		npm_package_version: string;
		npm_lifecycle_event: string;
		VIRTUAL_ENV: string;
		npm_config_globalconfig: string;
		npm_config_init_module: string;
		PWD: string;
		npm_execpath: string;
		npm_config_global_prefix: string;
		HOMEBREW_REPOSITORY: string;
		npm_command: string;
		TZ: string;
		VTE_VERSION: string;
		START_DOCKER: string;
		GOOGLE_API_KEY: string;
		OPENCLAW_PATH_BOOTSTRAPPED: string;
		INIT_CWD: string;
		EDITOR: string;
		NODE_ENV: string;
		[key: `PUBLIC_${string}`]: undefined;
		[key: `${string}`]: string | undefined;
	}
}

/**
 * Similar to [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private), but only includes variables that begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) (which defaults to `PUBLIC_`), and can therefore safely be exposed to client-side code.
 * 
 * Note that public dynamic environment variables must all be sent from the server to the client, causing larger network requests — when possible, use `$env/static/public` instead.
 * 
 * ```ts
 * import { env } from '$env/dynamic/public';
 * console.log(env.PUBLIC_DEPLOYMENT_SPECIFIC_VARIABLE);
 * ```
 */
declare module '$env/dynamic/public' {
	export const env: {
		[key: `PUBLIC_${string}`]: string | undefined;
	}
}
