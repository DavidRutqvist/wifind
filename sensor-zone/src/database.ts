'use strict';
import { ISensorLocationModel } from './models/sensor-location';
import * as Rx from 'rxjs/Rx';
import * as mongoose from 'mongoose';
import { ModelImpl } from "./model-impl";
import { ISensorLocation } from "./interfaces/sensor-location";
import { ISensorLocationResult } from "./interfaces/sensor-location-result";
import { SensorNotFoundError } from "./errors/sensor-not-found-error";
import { ZoneNotFoundError } from './errors/zone-not-found-error';

export class Database {
    private static _instance: Database;
    private readonly model: ModelImpl;

    constructor(connectionString: string) {
        let connection: mongoose.Connection = mongoose.createConnection(connectionString);
        this.model = new ModelImpl(connection);

        // use q promises which we will then wrap in rxjs's observable
        global.Promise = require("q").Promise;
        (<any>mongoose).Promise = global.Promise;
    }

    public static initialize(connectionString: string): void {
        Database._instance = new Database(connectionString);
    }

    public static getInstance(): Database {
        return Database._instance;
    }

    public getSensor(sensorId: string): Rx.Observable<ISensorLocationResult> {
        return Rx.Observable.fromPromise(this.model.sensorLocation.find({
            sensorId: sensorId,
        },
        {_id: false, __v: false})
        .sort({from: -1})
        .limit(1))
        .flatMap(locations => this.throwSensorNotFound(locations))
        .flatMap(x => x)
        .first()
        .map(location => this.toResult(location));
    }

    public getLocations(sensorId: string): Rx.Observable<ISensorLocationResult[]> {
        return Rx.Observable.fromPromise(this.model.sensorLocation.find({
            sensorId: sensorId
        },
        {_id: false, __v: false})
        .sort({from: -1}))
        .flatMap(locations => this.throwSensorNotFound(locations))
        .flatMap(x => x)
        .map(location => this.toResult(location))
        .toArray();
    }

    public getZoneSensors(zoneId: string): Rx.Observable<ISensorLocationResult[]> {
        return Rx.Observable.fromPromise(this.model.sensorLocation.find({
            zoneId: zoneId
        },
        {_id: false, __v: false})
        .sort({from: -1}))
        .flatMap(locations => this.throwZoneNotFound(locations))
        .flatMap(x => x)
        .map(location => this.toResult(location))
        .toArray();
    }
    
    public updateLastIfNull(sensorId: string, from: Date): Rx.Observable<ISensorLocation> {
        return Rx.Observable.fromPromise(this.model.sensorLocation.find({
            sensorId: sensorId
        },
        {__v: false})
        .sort({from: -1})
        .limit(1))
        .elementAt(0,null)
        .flatMap(x => this.setIfNull(x[0], from));
    }

    private setIfNull(latest: ISensorLocationModel, from: Date): Rx.Observable<ISensorLocation> {
        if (latest && !latest.to) {
            latest.to = from;
            return Rx.Observable.fromPromise(latest.save());
        } else {
            return Rx.Observable.of(latest);
        }
    }

    public addSensorLocation(sensorId: string, zoneId: string, from: Date, to?: Date): Rx.Observable<ISensorLocationResult> {
        
        const location: ISensorLocation = {
            sensorId: sensorId,
            zoneId: zoneId,
            from: from,
            to: to
        };

        return Rx.Observable.fromPromise(new this.model.sensorLocation(location).save())
            .map(location => this.toResult(location));
    }

    private throwSensorNotFound(locations: ISensorLocation[]): Rx.Observable<ISensorLocation[]> {
        if (locations && locations.length > 0) {
            return Rx.Observable.of(locations);
        } else {
            return Rx.Observable.throw(new SensorNotFoundError("The specified sensor could not be found"));
        }
    }

    private throwZoneNotFound(locations: ISensorLocation[]): Rx.Observable<ISensorLocation[]> {
        if (locations && locations.length > 0) {
            return Rx.Observable.of(locations);
        } else {
            return Rx.Observable.throw(new ZoneNotFoundError("The specified zone could not be found"));
        }
    }

    private toResult(location: ISensorLocation): ISensorLocationResult {
        return {
            sensorId: location.sensorId,
            zoneId: location.zoneId,
            from: location.from.getTime() / 1000,
            to: location.to ? location.to.getTime() / 1000 : undefined
        };
    }
}