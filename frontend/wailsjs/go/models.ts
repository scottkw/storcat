export namespace config {
	
	export class Config {
	    theme: string;
	    sidebarPosition: string;
	    windowWidth: number;
	    windowHeight: number;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.theme = source["theme"];
	        this.sidebarPosition = source["sidebarPosition"];
	        this.windowWidth = source["windowWidth"];
	        this.windowHeight = source["windowHeight"];
	    }
	}

}

export namespace models {
	
	export class CatalogMetadata {
	    title: string;
	    name: string;
	    filename: string;
	    created: string;
	    modified: string;
	    path: string;
	    hasHtml: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CatalogMetadata(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.name = source["name"];
	        this.filename = source["filename"];
	        this.created = source["created"];
	        this.modified = source["modified"];
	        this.path = source["path"];
	        this.hasHtml = source["hasHtml"];
	    }
	}
	export class SearchResult {
	    catalog: string;
	    catalogFilePath: string;
	    basename: string;
	    fullPath: string;
	    fullName: string;
	    type: string;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new SearchResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.catalog = source["catalog"];
	        this.catalogFilePath = source["catalogFilePath"];
	        this.basename = source["basename"];
	        this.fullPath = source["fullPath"];
	        this.fullName = source["fullName"];
	        this.type = source["type"];
	        this.size = source["size"];
	    }
	}

}

