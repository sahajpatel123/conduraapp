export namespace main {
	
	export class DaemonStatusResult {
	    ready: boolean;
	    addr: string;
	
	    static createFrom(source: any = {}) {
	        return new DaemonStatusResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ready = source["ready"];
	        this.addr = source["addr"];
	    }
	}

}

