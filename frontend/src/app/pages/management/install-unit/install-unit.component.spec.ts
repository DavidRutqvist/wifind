import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { InstallUnitComponent } from './install-unit.component';

describe('InstallUnitComponent', () => {
  let component: InstallUnitComponent;
  let fixture: ComponentFixture<InstallUnitComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ InstallUnitComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(InstallUnitComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should be created', () => {
    expect(component).toBeTruthy();
  });
});
