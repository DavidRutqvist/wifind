'use strict';

export class ZoneNotFoundError extends Error {
    constructor(message: string) {
        super(message);
    }
}