import { Component, OnInit } from '@angular/core';
import { SensorsService } from 'app/shared/services/sensors/sensors.service';
import { Zone } from 'app/shared/services/zones/zone';
import { ZonesService } from 'app/shared/services/zones/zones.service';
import swal from "sweetalert2";

@Component({
  selector: 'app-install-unit',
  templateUrl: './install-unit.component.html',
  styleUrls: ['./install-unit.component.scss']
})
export class InstallUnitComponent implements OnInit {
  selectedUnit: string;
  isUnitsLoading = false;
  units: string[] = [];

  selectedZone: string;
  isZonesLoading = false;
  zones: Zone[] = [];

  isSaving = false;

  constructor(
    private readonly sensorsSvc: SensorsService,
    private readonly zonesSvc: ZonesService) { }

  ngOnInit() {
    this.loadSensors();
    this.loadZones();
  }

  loadZones(): void {
    this.isZonesLoading = true;
    this.zonesSvc.getAllZones()
      .finally(() => this.isZonesLoading = false)
      .subscribe(zones => this.zones = zones);
  }

  loadSensors(): void {
    this.isUnitsLoading = true;
    this.sensorsSvc.getSensors()
      .finally(() => this.isUnitsLoading = false)
      .subscribe(units => this.units = units);
  }

  reset(): void {
    this.selectedUnit = "";
    this.selectedZone = "";
  }

  install(): void {
    if (this.selectedUnit && this.selectedZone) {
      this.isSaving = true;
      this.zonesSvc.addSensorToZone(this.selectedZone, this.selectedUnit)
        .finally(() => this.isSaving = false)
        .subscribe(
          () => this.installSuccessful(),
          err => swal("Ouch!", err.message, "error").catch(swal.noop)
        );
    }
  }

  installSuccessful(): void {
    swal("Nice!", "Unit successfully installed", "success").catch(swal.noop);
    this.reset();
  }

}
