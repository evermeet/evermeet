export function base64ToBuffer(base64: string): ArrayBuffer {
	try {
		let str = base64.replace(/-/g, '+').replace(/_/g, '/');
		while (str.length % 4) str += '=';
		const binary = window.atob(str);
		const bytes = new Uint8Array(binary.length);
		for (let i = 0; i < binary.length; i++) {
			bytes[i] = binary.charCodeAt(i);
		}
		return bytes.buffer;
	} catch (e) {
		console.error('Failed to decode base64:', base64, e);
		throw e;
	}
}

export function bufferToBase64(buffer: ArrayBuffer): string {
	const bytes = new Uint8Array(buffer);
	let binary = '';
	for (let i = 0; i < bytes.byteLength; i++) {
		binary += String.fromCharCode(bytes[i]);
	}
	return window.btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
}

// recursive conversion for WebAuthn options
export function recursiveBase64ToBuffer(obj: any, parentKey?: string): any {
	if (!obj || typeof obj !== 'object') return obj;
	
	if (Array.isArray(obj)) {
		return obj.map(v => recursiveBase64ToBuffer(v, parentKey));
	}

	for (const key in obj) {
		const val = obj[key];
		if (typeof val === 'string') {
			// challenge and userHandle are always base64
			if (key === 'challenge' || key === 'userHandle') {
				obj[key] = base64ToBuffer(val);
			}
			// id is binary for users and credentials, but NOT for rp
			if (key === 'id' && parentKey !== 'rp') {
				obj[key] = base64ToBuffer(val);
			}
		} else if (typeof val === 'object') {
			obj[key] = recursiveBase64ToBuffer(val, key);
		}
	}
	return obj;
}
