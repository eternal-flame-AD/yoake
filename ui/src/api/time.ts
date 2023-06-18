export function format_rust_naive_date(date: Date): string {
    const rfc3339 = date.toISOString();

    // Rust's NaiveDateTime::parse_from_rfc3339() doesn't like the trailing 'Z'
    // that JS's toISOString() adds, so we strip it off.
    return rfc3339.slice(0, rfc3339.length - 1);

}

export function parse_rust_naive_date(date: string): Date {
    return new Date(Date.parse(date + "Z"));
}