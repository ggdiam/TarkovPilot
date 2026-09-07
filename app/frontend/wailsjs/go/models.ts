export namespace api {
	
	export class QuestImportResult {
	    charUid: string;
	    importedCount: number;
	    alreadyDone: number;
	    ignoredCount: number;
	
	    static createFrom(source: any = {}) {
	        return new QuestImportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.charUid = source["charUid"];
	        this.importedCount = source["importedCount"];
	        this.alreadyDone = source["alreadyDone"];
	        this.ignoredCount = source["ignoredCount"];
	    }
	}
	export class QuestSyncCharacter {
	    uid: string;
	    charName: string;
	    gameMode: string;
	    fraction: string;
	    level: number;
	
	    static createFrom(source: any = {}) {
	        return new QuestSyncCharacter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.uid = source["uid"];
	        this.charName = source["charName"];
	        this.gameMode = source["gameMode"];
	        this.fraction = source["fraction"];
	        this.level = source["level"];
	    }
	}
	export class QuestSyncProfile {
	    profileKey: string;
	    gameMode: string;
	    foundCount: number;
	    matchedCount: number;
	    ignoredCount: number;
	    firstEventAt: string;
	    lastEventAt: string;
	    linkedCharUid: string;
	
	    static createFrom(source: any = {}) {
	        return new QuestSyncProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.profileKey = source["profileKey"];
	        this.gameMode = source["gameMode"];
	        this.foundCount = source["foundCount"];
	        this.matchedCount = source["matchedCount"];
	        this.ignoredCount = source["ignoredCount"];
	        this.firstEventAt = source["firstEventAt"];
	        this.lastEventAt = source["lastEventAt"];
	        this.linkedCharUid = source["linkedCharUid"];
	    }
	}
	export class QuestSyncResponse {
	    ok: boolean;
	    error: string;
	    profiles: QuestSyncProfile[];
	    characters: QuestSyncCharacter[];
	    skippedEvents: number;
	    sourceSince: string;
	    updatedAt: string;
	    importResult?: QuestImportResult;
	
	    static createFrom(source: any = {}) {
	        return new QuestSyncResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.error = source["error"];
	        this.profiles = this.convertValues(source["profiles"], QuestSyncProfile);
	        this.characters = this.convertValues(source["characters"], QuestSyncCharacter);
	        this.skippedEvents = source["skippedEvents"];
	        this.sourceSince = source["sourceSince"];
	        this.updatedAt = source["updatedAt"];
	        this.importResult = this.convertValues(source["importResult"], QuestImportResult);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

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
	    pro: boolean;
	    connState: string;
	    autoStart: boolean;
	    autoClean: boolean;
	    startMinimized: boolean;
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
	        this.pro = source["pro"];
	        this.connState = source["connState"];
	        this.autoStart = source["autoStart"];
	        this.autoClean = source["autoClean"];
	        this.startMinimized = source["startMinimized"];
	        this.eventLog = source["eventLog"];
	    }
	}

}

