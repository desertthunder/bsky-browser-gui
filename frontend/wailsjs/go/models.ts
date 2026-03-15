export namespace main {
	
	export class Auth {
	    did: string;
	    handle: string;
	    access_jwt: string;
	    refresh_jwt: string;
	    pds_url: string;
	    session_id: string;
	    auth_server_url: string;
	    auth_server_token_endpoint: string;
	    auth_server_revocation_endpoint: string;
	    dpop_auth_nonce: string;
	    dpop_host_nonce: string;
	    dpop_private_key: string;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Auth(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.did = source["did"];
	        this.handle = source["handle"];
	        this.access_jwt = source["access_jwt"];
	        this.refresh_jwt = source["refresh_jwt"];
	        this.pds_url = source["pds_url"];
	        this.session_id = source["session_id"];
	        this.auth_server_url = source["auth_server_url"];
	        this.auth_server_token_endpoint = source["auth_server_token_endpoint"];
	        this.auth_server_revocation_endpoint = source["auth_server_revocation_endpoint"];
	        this.dpop_auth_nonce = source["dpop_auth_nonce"];
	        this.dpop_host_nonce = source["dpop_host_nonce"];
	        this.dpop_private_key = source["dpop_private_key"];
	        this.updated_at = this.convertValues(source["updated_at"], null);
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
	export class LogEntry {
	    level: string;
	    message: string;
	    // Go type: time
	    timestamp: any;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.level = source["level"];
	        this.message = source["message"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
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
	export class SearchResult {
	    uri: string;
	    cid: string;
	    author_did: string;
	    author_handle: string;
	    text: string;
	    // Go type: time
	    created_at: any;
	    like_count: number;
	    repost_count: number;
	    reply_count: number;
	    source: string;
	    facets: string;
	    // Go type: time
	    indexed_at: any;
	    rank: number;
	
	    static createFrom(source: any = {}) {
	        return new SearchResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.uri = source["uri"];
	        this.cid = source["cid"];
	        this.author_did = source["author_did"];
	        this.author_handle = source["author_handle"];
	        this.text = source["text"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.like_count = source["like_count"];
	        this.repost_count = source["repost_count"];
	        this.reply_count = source["reply_count"];
	        this.source = source["source"];
	        this.facets = source["facets"];
	        this.indexed_at = this.convertValues(source["indexed_at"], null);
	        this.rank = source["rank"];
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

