'use strict';
Object.defineProperty(exports, "__esModule", { value: true });
const Rx = require("rxjs/Rx");
const mongoose = require("mongoose");
const model_impl_1 = require("./model-impl");
const sensor_not_found_error_1 = require("./errors/sensor-not-found-error");
class Database {
    constructor(connectionString) {
        let connection = mongoose.createConnection(connectionString);
        this.model = new model_impl_1.ModelImpl(connection);
        global.Promise = require("q").Promise;
        mongoose.Promise = global.Promise;
    }
    static initialize(connectionString) {
        Database._instance = new Database(connectionString);
    }
    static getInstance() {
        return Database._instance;
    }
    getSensor(sensorId) {
        return Rx.Observable.fromPromise(this.model.sensorLocation.find({
            sensorId: sensorId,
        }, { _id: false, __v: false })
            .sort({ from: -1 })
            .limit(1))
            .flatMap(locations => this.throwSensorNotFound(locations))
            .flatMap(x => x)
            .first()
            .map(location => this.toResult(location));
    }
    getLocations(sensorId) {
        return Rx.Observable.fromPromise(this.model.sensorLocation.find({
            sensorId: sensorId
        }, { _id: false, __v: false })
            .sort({ from: -1 }))
            .flatMap(locations => this.throwSensorNotFound(locations))
            .flatMap(x => x)
            .map(location => this.toResult(location))
            .toArray();
    }
    getZoneSensors(zoneId) {
        return Rx.Observable.fromPromise(this.model.sensorLocation.find({
            zoneId: zoneId
        }, { _id: false, __v: false })
            .sort({ from: -1 }))
            .flatMap(locations => this.throwSensorNotFound(locations))
            .flatMap(x => x)
            .map(location => this.toResult(location))
            .toArray();
    }
    updateLastIfNull(sensorId, from) {
        return Rx.Observable.fromPromise(this.model.sensorLocation.find({
            sensorId: sensorId
        }, { __v: false })
            .sort({ from: -1 })
            .limit(1))
            .elementAt(0, null)
            .flatMap(x => this.setIfNull(x[0], from));
    }
    setIfNull(latest, from) {
        if (latest && !latest.to) {
            latest.to = from;
            return Rx.Observable.fromPromise(latest.save());
        }
        else {
            return Rx.Observable.of(latest);
        }
    }
    addSensorLocation(sensorId, zoneId, from, to) {
        const location = {
            sensorId: sensorId,
            zoneId: zoneId,
            from: from,
            to: to
        };
        return Rx.Observable.fromPromise(new this.model.sensorLocation(location).save())
            .map(location => this.toResult(location));
    }
    throwSensorNotFound(locations) {
        if (locations && locations.length > 0) {
            return Rx.Observable.of(locations);
        }
        else {
            return Rx.Observable.throw(new sensor_not_found_error_1.SensorNotFoundError("The specified sensor could not be found"));
        }
    }
    toResult(location) {
        return {
            sensorId: location.sensorId,
            zoneId: location.zoneId,
            from: location.from.getTime() / 1000,
            to: location.to ? location.to.getTime() / 1000 : undefined
        };
    }
}
exports.Database = Database;
