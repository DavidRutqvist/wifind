'use strict';

import * as mongoose from "mongoose";

// Models
import { IModel } from "./models/model";
import { ISensorLocationModel } from "./models/sensor-location";

// Schemas
import { sensorLocationSchema } from "./schemas/sensor-location";

export class ModelImpl implements IModel {
    readonly sensorLocation: mongoose.Model<ISensorLocationModel>;

    constructor(connection: mongoose.Connection) {
        this.sensorLocation = connection.model<ISensorLocationModel>("SensorLocation", sensorLocationSchema);
    }
}