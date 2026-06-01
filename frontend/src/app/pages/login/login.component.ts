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
  selector: 'app-login',
  imports: [
    CommonModule,
    ReactiveFormsModule,
  ],
  templateUrl: './login.component.html',
  styleUrl: './login.component.css'
})
export class LoginComponent {

  loading = false;

  errorMessage = '';

  loginForm;

  constructor(
    private fb: FormBuilder,
    private authService: AuthService,
    private router: Router,
  ) {

    this.loginForm = this.fb.group({

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
        ],
      ],
    });
  }

  onSubmit(): void {

    if (this.loginForm.invalid) {
      return;
    }

    this.loading = true;

    this.authService.login(
      this.loginForm.value
    ).subscribe({

      next: (response) => {

        this.loading = false;

        this.authService.saveToken(
          response.token
        );

        this.router.navigate(['/dashboard']);
      },

      error: (err) => {

        this.loading = false;

        this.errorMessage =
          err.error?.error || 'Login failed';
      },
    });
  }
}