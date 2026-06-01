import { Injectable } from '@angular/core';

import { HttpClient } from '@angular/common/http';

import { Observable } from 'rxjs';

import { environment } from '../../environments/environment';

@Injectable({
  providedIn: 'root',
})
export class AuthService {

  private apiUrl = environment.apiUrl;

  constructor(private http: HttpClient) {}

  signup(data: any): Observable<any> {

    return this.http.post(
      `${this.apiUrl}/signup`,
      data
    );
  }

  login(data: any): Observable<any> {

    return this.http.post(
      `${this.apiUrl}/login`,
      data
    );
  }

  saveToken(token: string): void {

    localStorage.setItem('token', token);
  }

  getToken(): string | null {

    return localStorage.getItem('token');
  }

  logout(): void {

    localStorage.removeItem('token');
  }

  isAuthenticated(): boolean {

    return !!this.getToken();
  }
}