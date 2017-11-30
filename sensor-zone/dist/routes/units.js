"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const sensor_not_found_error_1 = require("../errors/sensor-not-found-error");
const log = require("winston");
const database_1 = require("../database");
class UnitsRoute {
    static create(router) {
        log.info("[UnitsRoute::create] Creating units route.");
        router.get("/sensors/:sensorid", (req, res, next) => {
            new UnitsRoute().getSensor(req, res, next);
        });
        router.get("/sensors/:sensorid/zones", (req, res, next) => {
            new UnitsRoute().getZones(req, res, next);
        });
        router.get("/sensors/zones/:zoneId", (req, res, next) => {
            new UnitsRoute().getZoneSensors(req, res, next);
        });
        router.post("/sensors", (req, res, next) => {
            new UnitsRoute().addSensor(req, res, next);
        });
        router.put("/sensors/:sensorid", (req, res, next) => {
            new UnitsRoute().updateSensor(req, res, next);
        });
        router.delete("/sensors/:sensorid", (req, res, next) => {
            new UnitsRoute().deleteSensor(req, res, next);
        });
    }
    constructor() { }
    getSensor(req, res, next) {
        database_1.Database.getInstance().getSensor(req.params.sensorid)
            .subscribe(sensor => res.json({
            success: true,
            sensor: sensor
        }), err => this.logAndReturn(res, err));
    }
    getZoneSensors(req, res, next) {
        database_1.Database.getInstance().getZoneSensors(req.params.zoneId)
            .subscribe(sensors => res.json({
            success: true,
            sensors: sensors
        }), err => this.logAndReturn(res, err));
    }
    getZones(req, res, next) {
        database_1.Database.getInstance().getLocations(req.params.sensorid)
            .subscribe(locations => res.json({
            success: true,
            locations: locations
        }), err => this.logAndReturn(res, err));
    }
    addSensor(req, res, next) {
        if (!req.body.zoneId) {
            res.status(400).json({
                success: false,
                message: "Missing zoneId in body"
            });
            return;
        }
        if (!req.body.sensorId) {
            res.status(400).json({
                success: false,
                message: "Missing sensorId in body"
            });
            return;
        }
        if (!req.body.from) {
            res.status(400).json({
                success: false,
                message: "Missing from in body"
            });
            return;
        }
        const fromTimestamp = req.body.from;
        const from = new Date(fromTimestamp * 1000);
        let to = undefined;
        if (req.body.to) {
            const toTimestamp = req.body.to;
            to = new Date(toTimestamp * 1000);
        }
        database_1.Database.getInstance().updateLastIfNull(req.body.sensorId, from)
            .flatMap(() => database_1.Database.getInstance().addSensorLocation(req.body.sensorId, req.body.zoneId, from, to))
            .subscribe(location => res.json({
            success: true,
            location: location
        }), err => this.logAndReturn(res, err));
    }
    updateSensor(req, res, next) {
        res.status(400).json({
            success: false,
            message: "Not yet implemented"
        });
    }
    deleteSensor(req, res, next) {
        res.status(400).json({
            success: false,
            message: "Not yet implemented"
        });
    }
    logAndReturn(res, err) {
        if (err instanceof sensor_not_found_error_1.SensorNotFoundError) {
            res.status(404).json({
                success: false,
                message: "The sensor could not be found"
            });
        }
        else {
            log.error("Error in API", err);
            res.status(500).json({
                success: false,
                message: "Something went wrong"
            });
        }
    }
}
exports.UnitsRoute = UnitsRoute;
