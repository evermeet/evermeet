
import Package from '../../../../package.json';
import { internalApiCall } from './api.js';

let loadedConfig;

// load configuration
export async function loadConfig () {
    if (!loadedConfig) {
        loadedConfig = await internalApiCall(fetch, '_config');
    }
    return loadedConfig
}

export const pkg = Package;