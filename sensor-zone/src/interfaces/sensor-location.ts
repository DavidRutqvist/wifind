'use strict';

export interface ISensorLocation {
    sensorId: string;
    zoneId: string;
    from: Date;
    to?: Date;
}
