import { Component, OnInit } from '@angular/core';

@Component({
  selector: 'app-install-unit',
  templateUrl: './install-unit.component.html',
  styleUrls: ['./install-unit.component.scss']
})
export class InstallUnitComponent implements OnInit {
  selectedUnit: string;
  isUnitsLoading = false;
  units: string[] = ["hej", "då"];

  constructor() { }

  ngOnInit() {
  }

}
