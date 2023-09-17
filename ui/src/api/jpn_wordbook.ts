import { makeJSONRequest } from "./request";

export interface LookupResult {
    ja: string,
    altn: string[] | null,
    jm: string[] | null,
    fu: string | null,
    en: string[] | null,
    ex: string[] | null,

    src: string,
}

export interface WordbookItem {
    uuid: string,

    ja: string,
    altn: string[],
    jm: string[],
    fu: string,
    en: string[],
    ex: string[],

    src: string,

    created: string,
    updated: string,
}

export async function comboSearchWord(query: string): Promise<LookupResult[]> {
    const url = "/api/jpn_wordbook/sources/combo/search?query=" + encodeURIComponent(query);
    const method = "GET";
    let ret = (await makeJSONRequest<LookupResult[]>(url, method));
    if (ret.status != "Ok") {
        throw new Error(ret.message);
    }
    return ret.data;
}

export async function comboSearchWordTop(query: string): Promise<LookupResult> {
    const url = "/api/jpn_wordbook/sources/combo/search_top?query=" + encodeURIComponent(query);
    const method = "GET";
    let ret = (await makeJSONRequest<LookupResult>(url, method));
    if (ret.status != "Ok") {
        throw new Error(ret.message);
    }
    return ret.data;
}

export async function storeWordbook(word: LookupResult): Promise<void> {
    const url = "/api/jpn_wordbook/wordbook";
    const method = "POST";
    const body = word;
    let ret = (await makeJSONRequest<void>(url, method, body));
    if (ret.status != "Ok") {
        throw new Error(ret.message);
    }
}

export interface getWordbookParams {
    until?: string
    limit: number,
}

export async function getWordbook(params: getWordbookParams): Promise<WordbookItem[]> {
    let url = `/api/jpn_wordbook/wordbook?limit=${params.limit}`;
    if (params.until) {
        url += `&until=${params.until}`;
    }
    const method = "GET";
    let ret = (await makeJSONRequest<WordbookItem[]>(url, method));
    if (ret.status != "Ok") {
        throw new Error(ret.message);
    }
    return ret.data;
}

export function downloadWordbookCsv(header: boolean) {
    let url = `/api/jpn_wordbook/wordbook/csv_export?header=${header}`;
    let link = document.createElement("a");
    link.download = "wordbook.csv";
    link.href = url;
    link.click();
}