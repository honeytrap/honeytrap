import axios from 'axios';

export const RECEIVED_EVENT = 'RECEIVED_EVENT';
export const RECEIVED_EVENTS = 'RECEIVED_EVENTS';
export const CONNECTION_STATUS = 'CONNECTION_STATUS';

export const CLEAR_HOT_COUNTRIES = 'CLEAR_HOT_COUNTRIES';
export const RECEIVED_HOT_COUNTRIES = 'RECEIVED_HOT_COUNTRIES';
export const RECEIVED_METADATA = 'RECEIVED_METADATA';

export const FETCH_COUNTRIES = 'FETCH_COUNTRIES';

export const ADD_SESSION = 'ADD_SESSION';
export const FETCH_SESSIONS = 'FETCH_SESSIONS';
export const FETCH_SESSION = 'FETCH_SESSION';
export const FETCH_SESSION_CONTENT = 'FETCH_SESSION_CONTENT';

const ROOT_URL = 'http://127.0.0.1:8089';

export function clearHotCountries(event) {
	  return {
		    type: HOT_COUNTRIES,
        payload: {},
	  };
}

export function receivedHotCountries(event) {
	  return {
		    type: RECEIVED_HOT_COUNTRIES,
        payload: event
	  };
}

export function receivedMetadata(event) {
	  return {
		    type: RECEIVED_METADATA,
        payload: event
	  };
}

export function receivedEvents(data) {
	  return {
		    type: RECEIVED_EVENTS,
        payload: data
	  };
}

export function receivedEvent(event) {
	  return {
		    type: RECEIVED_EVENT,
        payload: event
	  };
}

export function connectionStatus(connected) {
	  return {
		    type: CONNECTION_STATUS,
		    payload: {
            connected: connected, 
        }
	  };
}

export function addSession(id) {
	return {
		type: ADD_SESSION,
		payload: {
			id: id, 
			date: '10/02/2017', 
			location: 'unknown', 
			started: '10/02/2017 10:10', 
			ended: '10/02/2017 10:11', 
			username: 'root', 
			password: 'root' 
		}
	};
}

export function fetchCountries() {
    const request = axios.all([axios.get(`https://unpkg.com/world-atlas@1/world/110m.json`), axios.get(`https://unpkg.com/world-atlas@1/world/110m.tsv`)]);

	  return {
		    type: FETCH_COUNTRIES,
		    payload: request
    };
}

export function fetchSessions() {
	const request = axios.get(`${ROOT_URL}/api/v1/sessions`);

	return {
		type: FETCH_SESSIONS,
		payload: request
	}
}

export function fetchSession(id) {
	const request = axios.get(`${ROOT_URL}/api/v1/sessions/${id}`);

	return {
		type: FETCH_SESSION,
		payload: request
	}
}

export function fetchSessionContent(content) {

	return {
		type: FETCH_SESSION_CONTENT,
		payload: content
	}
}
