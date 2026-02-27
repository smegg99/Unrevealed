// stealth.go
package unrevealed

// StealthScripts returns JavaScript snippets to inject via CDP
// Page.addScriptToEvaluateOnNewDocument for bot detection evasion.
func StealthScripts() []string {
	return []string{
		scriptWebdriver,
		scriptChrome,
		scriptPermissions,
		scriptFnToString,
	}
}

// StealthFlags returns Chrome command-line flags for stealth operation.
// Keys are flag names, values are flag values (empty string for boolean flags).
func StealthFlags() map[string]string {
	return map[string]string{
		"disable-blink-features":   "AutomationControlled",
		"no-first-run":             "",
		"no-default-browser-check": "",
		"no-sandbox":               "",
		"window-size":              "1920,1080",
		"start-maximized":          "",
		"disable-infobars":         "",
	}
}

// DeleteFlags returns Chrome flags that should be removed from launch
// arguments to avoid detection (typically added by automation libraries).
func DeleteFlags() []string {
	return []string{
		"enable-automation",
	}
}

// Override navigator.webdriver on the prototype.
// The --disable-blink-features=AutomationControlled flag handles this at the
// Chrome level, but this JS override acts as a fallback for edge cases.
var scriptWebdriver = `
Object.defineProperty(Navigator.prototype, 'webdriver', {
  get: () => undefined,
  configurable: true,
});
`

// Mock window.chrome.runtime which is absent in automated browsers
// but present in regular Chrome sessions.
var scriptChrome = `
if (!window.chrome) {
  window.chrome = {};
}
if (!window.chrome.runtime) {
  window.chrome.runtime = {
    OnInstalledReason: {
      CHROME_UPDATE: 'chrome_update',
      INSTALL: 'install',
      SHARED_MODULE_UPDATE: 'shared_module_update',
      UPDATE: 'update',
    },
    OnRestartRequiredReason: {
      APP_UPDATE: 'app_update',
      OS_UPDATE: 'os_update',
      PERIODIC: 'periodic',
    },
    PlatformArch: {
      ARM: 'arm',
      ARM64: 'arm64',
      MIPS: 'mips',
      MIPS64: 'mips64',
      X86_32: 'x86-32',
      X86_64: 'x86-64',
    },
    PlatformNaclArch: {
      ARM: 'arm',
      MIPS: 'mips',
      MIPS64: 'mips64',
      X86_32: 'x86-32',
      X86_64: 'x86-64',
    },
    PlatformOs: {
      ANDROID: 'android',
      CROS: 'cros',
      LINUX: 'linux',
      MAC: 'mac',
      OPENBSD: 'openbsd',
      WIN: 'win',
    },
    RequestUpdateCheckStatus: {
      NO_UPDATE: 'no_update',
      THROTTLED: 'throttled',
      UPDATE_AVAILABLE: 'update_available',
    },
  };
}
`

// Fix navigator.permissions.query behavior and Notification API
// which differ in automated browsers.
var scriptPermissions = `
if (!window.Notification) {
  window.Notification = { permission: 'denied' };
}
const originalQuery = window.navigator.permissions.query;
window.navigator.permissions.__proto__.query = (parameters) =>
  parameters.name === 'notifications'
    ? Promise.resolve({ state: Notification.permission })
    : originalQuery(parameters);
`

// Mock navigator.plugins as a proper PluginArray and fix related navigator properties.
// var scriptPlugins = `
// (() => {
//   const makePlugin = (obj) => {
//     const p = Object.create(Plugin.prototype);
//     Object.assign(p, obj);
//     p.length = 1;
//     p[0] = { type: 'application/pdf', suffixes: 'pdf', description: obj.description || '', enabledPlugin: p };
//     return p;
//   };

//   const plugins = [
//     makePlugin({ name: 'Chrome PDF Plugin', filename: 'internal-pdf-viewer', description: 'Portable Document Format' }),
//     makePlugin({ name: 'Chrome PDF Viewer', filename: 'mhjfbmdgcfjbbpaeojofohoefgiehjai', description: '' }),
//     makePlugin({ name: 'Native Client', filename: 'internal-nacl-plugin', description: '' }),
//   ];

//   const pluginArray = Object.create(PluginArray.prototype);
//   plugins.forEach((p, i) => { pluginArray[i] = p; });
//   Object.defineProperty(pluginArray, 'length', { get: () => plugins.length });
//   pluginArray.item = (i) => pluginArray[i];
//   pluginArray.namedItem = (name) => plugins.find(p => p.name === name);
//   pluginArray.refresh = () => {};
//   pluginArray[Symbol.iterator] = function* () { for (let i = 0; i < plugins.length; i++) yield pluginArray[i]; };

//   Object.defineProperty(navigator, 'plugins', { get: () => pluginArray });
// })();

// Object.defineProperty(navigator, 'languages', {
//   get: () => ['en-US', 'en'],
// });
// if (navigator.connection) {
//   Object.defineProperty(navigator.connection, 'rtt', { get: () => 100 });
// }
// Object.defineProperty(navigator, 'maxTouchPoints', { get: () => 1 });
// `

// Spoof Function.prototype.toString so overridden native functions
// still return "[native code]" when inspected by detection scripts.
var scriptFnToString = `
(() => {
  const nativeStr = Error.toString().replace(/Error/g, 'toString');
  const oldToString = Function.prototype.toString;
  const oldCall = Function.prototype.call;

  function call() {
    return oldCall.apply(this, arguments);
  }
  Function.prototype.call = call;

  function functionToString() {
    if (this === window.navigator.permissions.query) {
      return 'function query() { [native code] }';
    }
    if (this === functionToString) {
      return nativeStr;
    }
    return oldCall.call(oldToString, this);
  }
  Function.prototype.toString = functionToString;
})();
`
