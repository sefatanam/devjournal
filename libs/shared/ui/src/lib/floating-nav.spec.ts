import { ComponentFixture, TestBed } from '@angular/core/testing';
import { FloatingNav } from './floating-nav';

describe('FloatingNav', () => {
  let component: FloatingNav;
  let fixture: ComponentFixture<FloatingNav>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [FloatingNav],
    }).compileComponents();

    fixture = TestBed.createComponent(FloatingNav);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
