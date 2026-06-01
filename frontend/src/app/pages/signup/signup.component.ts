import { Component } from '@angular/core';

import {
  FormBuilder,
  ReactiveFormsModule,
  Validators
} from '@angular/forms';

import { Router } from '@angular/router';

import { CommonModule } from '@angular/common';

import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-signup',
  imports: [
    CommonModule,
    ReactiveFormsModule,
  ],
  templateUrl: './signup.component.html',
  styleUrl: './signup.component.css'
})
export class SignupComponent {

  loading = false;

  errorMessage = '';

  signupForm;

  constructor(
    private fb: FormBuilder,
    private authService: AuthService,
    private router: Router,
  ) {

    this.signupForm = this.fb.group({

      email: [
        '',
        [
          Validators.required,
          Validators.email,
        ],
      ],

      password: [
        '',
        [
          Validators.required,
          Validators.minLength(6),
        ],
      ],
    });
  }

  onSubmit(): void {
    console.log('signup clicked');

    console.log(this.signupForm.value);

    if (this.signupForm.invalid) {
      console.log('form invalid');
      return;
    }

    // if (this.signupForm.invalid) {
    //   return;
    // }

    this.loading = true;

    this.authService.signup(
      this.signupForm.value
    ).subscribe({

      next: () => {

        this.loading = false;

        this.router.navigate(['/login']);
      },

      error: (err) => {

        this.loading = false;

        this.errorMessage =
          err.error?.error || 'Signup failed';
      },
    });
  }
}