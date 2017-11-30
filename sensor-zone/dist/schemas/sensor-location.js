'use strict';
Object.defineProperty(exports, "__esModule", { value: true });
const mongoose_1 = require("mongoose");
exports.sensorLocationSchema = new mongoose_1.Schema({
    sensorId: { type: String, index: true, required: true },
    zoneId: { type: mongoose_1.Schema.Types.ObjectId, index: true, required: true },
    from: { type: Date, required: true },
    to: { type: Date, required: false }
});
exports.sensorLocationSchema.index({ sensorId: 1, from: 1 });
