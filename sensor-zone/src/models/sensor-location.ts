import { Document } from "mongoose";
import { ISensorLocation } from "../interfaces/sensor-location";

export interface ISensorLocationModel extends ISensorLocation, Document {

}
