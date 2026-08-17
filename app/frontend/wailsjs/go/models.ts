export namespace main {
	
	export class State {
	    version: string;
	    latestVersion: string;
	    updateReady: boolean;
	    hookId: string;
	    host: string;
	    lang: string;
	    gameFolder: string;
	    gameFolderOverride: string;
	    screenshotsFolder: string;
	    screenshotsOverride: string;
	    gameFound: boolean;
	    logsWatching: boolean;
	    screenshotsWatching: boolean;
	    connState: string;
	    autoStart: boolean;
	    autoClean: boolean;
	    eventLog: string[];
	
	    static createFrom(source: any = {}) {
	        return new State(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.latestVersion = source["latestVersion"];
	        this.updateReady = source["updateReady"];
	        this.hookId = source["hookId"];
	        this.host = source["host"];
	        this.lang = source["lang"];
	        this.gameFolder = source["gameFolder"];
	        this.gameFolderOverride = source["gameFolderOverride"];
	        this.screenshotsFolder = source["screenshotsFolder"];
	        this.screenshotsOverride = source["screenshotsOverride"];
	        this.gameFound = source["gameFound"];
	        this.logsWatching = source["logsWatching"];
	        this.screenshotsWatching = source["screenshotsWatching"];
	        this.connState = source["connState"];
	        this.autoStart = source["autoStart"];
	        this.autoClean = source["autoClean"];
	        this.eventLog = source["eventLog"];
	    }
	}

}

