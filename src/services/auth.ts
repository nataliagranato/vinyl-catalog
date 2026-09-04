import { v4 as uuidv4 } from 'uuid';
import { Env } from '../types';
import { JWTService } from './jwt';

export class AuthService {
  private jwtService: JWTService;

  constructor(private env: Env) {
    // Use JWT_SECRET from env or fallback
    const jwtSecret = this.env.JWT_SECRET || 'default-insecure-secret';
    this.jwtService = new JWTService(jwtSecret);
  }

  async login(username: string, password: string): Promise<string | null> {
    const adminPassword = this.env.ADMIN_PASSWORD || 'admin';
    
    if (username !== this.env.ADMIN_USERNAME || password !== adminPassword) {
      return null;
    }

    const token = await this.jwtService.sign(
      { username },
      parseInt(this.env.JWT_EXPIRATION_HOURS) * 3600
    );

    return token;
  }

  async verifyToken(token: string): Promise<any> {
    return await this.jwtService.verify(token);
  }
}