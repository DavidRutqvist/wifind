'use strict';

export class SensorNotFoundError extends Error {
    constructor(message: string) {
        super(message);
    }
}