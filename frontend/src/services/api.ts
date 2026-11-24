import type { endpointMap, User } from "./types";

class Api {
    private baseUrl: string;
    private endpoints: endpointMap

    constructor(){
        this.baseUrl = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';
        this.endpoints = {
            register: "/register",
            login: "/login"
        }
    }


    public async register(username: string, password: string): Promise<User> {
        const url: string = this.baseUrl + this.endpoints.register
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            credentials: 'include',
            body: JSON.stringify({
                username: username,
                password: password
            })
        })

        if (!response.ok) {
            throw new Error(`Registration failed: ${response.status}`)
        }

        return response.json() as Promise<User>
    }

    public async login(username: string, password: string): Promise<User> {
        const url: string = this.baseUrl + this.endpoints.login
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            credentials: 'include',
            body: JSON.stringify({
                username: username,
                password: password
            })
        })

        if (!response.ok) {
            throw new Error(`Registration failed: ${response.status}`)
        }

        return response.json() as Promise<User>
    }
}
