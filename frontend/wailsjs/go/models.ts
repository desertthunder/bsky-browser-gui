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

}

