'use strict';
Object.defineProperty(exports, "__esModule", { value: true });
const sensor_location_1 = require("./schemas/sensor-location");
class ModelImpl {
    constructor(connection) {
        this.sensorLocation = connection.model("SensorLocation", sensor_location_1.sensorLocationSchema);
    }
}
exports.ModelImpl = ModelImpl;
