import { makeJSONRequest } from "./request";

export interface Medication {
    uuid: string,
    name: string,
    dosage: number,
    dosage_unit: string,
    period_hours: number,
    flags: string,
    options: string,
}

export async function parseShorthand(shorthand: string): Promise<Medication> {
    const url = "/api/med/parse_shorthand";
    const method = "POST";
    const body = {
        shorthand: shorthand,
    };
    return (await makeJSONRequest<Medication>(url, method, body)).data;
}

export async function formatShorthand(med: Medication): Promise<string> {
    const url = "/api/med/format_shorthand";
    const method = "POST";
    const body = med;
    return (await makeJSONRequest<string>(url, method, body)).data;
}

export async function getDirectives(): Promise<Medication[]> {
    const url = "/api/med/directive";
    const method = "GET";
    return (await makeJSONRequest<Medication[]>(url, method)).data;
}

export async function postDirective(med: Medication): Promise<Medication> {
    const url = "/api/med/directive";
    const method = "POST";
    const body = med;
    return (await makeJSONRequest<Medication>(url, method, body)).data;
}

export async function patchDirective(med: Medication): Promise<Medication> {
    const url = "/api/med/directive";
    const method = "PATCH";
    const body = med;

    return (await makeJSONRequest<Medication>(url, method, body)).data;
}

export async function deleteDirective(uuid: string): Promise<void> {
    const url = "/api/med/directive/" + uuid;
    const method = "DELETE";

    await makeJSONRequest<void>(url, method);
}