'use strict';

export interface SensorLocation {
    sensorId: string;
    zoneId: string;
    from: number;
    to?: number;
}
