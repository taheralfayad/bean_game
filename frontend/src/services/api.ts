import { endpointMap } from "./types";

class Api {
    private baseUrl: string;
    private endpoints: endpointMap

    constructor(){
        this.baseUrl = `https://${window.location.hostname}`;
        this.endpoints = {
            createGame: '/games'
        }
    }
}