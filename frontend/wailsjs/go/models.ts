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
	    size: number;
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
	        this.size = source["size"];
	        this.created = source["created"];
	        this.modified = source["modified"];
	        this.path = source["path"];
	        this.hasHtml = source["hasHtml"];
	    }
	}
	export class CreateCatalogResult {
	    jsonPath: string;
	    htmlPath: string;
	    fileCount: number;
	    totalSize: number;
	    copyJsonPath?: string;
	    copyHtmlPath?: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateCatalogResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.jsonPath = source["jsonPath"];
	        this.htmlPath = source["htmlPath"];
	        this.fileCount = source["fileCount"];
	        this.totalSize = source["totalSize"];
	        this.copyJsonPath = source["copyJsonPath"];
	        this.copyHtmlPath = source["copyHtmlPath"];
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

