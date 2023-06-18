

type APIStatus = "Ok" | "Error";

type APIResponse<T> = {
    code: number;
    status: APIStatus,
    message: string,
    data: T,

}

export async function makeJSONRequest<T>(url: string, method: string, body?: any): Promise<APIResponse<T>> {
    const response = await fetch(url, {
        method: method,
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify(body)
    });
    const data = await response.json();
    return data;
}