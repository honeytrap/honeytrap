import { combineReducers } from 'redux';
import SessionReducer from './reducer_sessions';

const rootReducer = combineReducers({
	sessions: SessionReducer,
});

export default rootReducer;