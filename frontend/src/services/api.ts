import type { endpointMap } from "./types";

class Api {
    private baseUrl: string;
    private endpoints: endpointMap

    constructor(){
        this.baseUrl = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';
        this.endpoints = {}
    }
}
