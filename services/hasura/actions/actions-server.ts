import express from 'express';
import bodyParser from 'body-parser';
import cors from 'cors';
// Zod is only used in handlers files, not here
import calculatePregnancyRisk from './handlers/calculatePregnancyRisk';
import findNearbyFacilities from './handlers/findNearbyFacilities';
import dispatchAmbulance from './handlers/dispatchAmbulance';
import scheduleAppointment from './handlers/scheduleAppointment';

// Create express app
const app = express();
const PORT = process.env.PORT || 3000;

// Setup middleware
app.use(bodyParser.json());
app.use(cors());

// Enable structured logging
const logger = {
  info: (message: string, meta?: Record<string, unknown>): void => {
    if (process.env.NODE_ENV !== 'production') {
      console.log(`[INFO] ${message}`, meta || '');
    } else {
      // In production, we would use a proper logging service
      console.log(JSON.stringify({
        level: 'info',
        message,
        timestamp: new Date().toISOString(),
        ...meta
      }));
    }
  },
  error: (message: string, error?: unknown, meta?: Record<string, unknown>): void => {
    const errorMessage = error instanceof Error ? error.message : String(error);
    if (process.env.NODE_ENV !== 'production') {
      console.error(`[ERROR] ${message}`, errorMessage, meta || '');
    } else {
      console.error(JSON.stringify({
        level: 'error',
        message,
        error: errorMessage,
        timestamp: new Date().toISOString(),
        ...meta
      }));
    }
  }
};

// Validate JWT token (when authentication is enabled)
const validateJwt = (_req: express.Request, res: express.Response, next: express.NextFunction): void => {
  try {
    // Extract JWT from Authorization header
    //const authHeader = req.headers.authorization;
    
    // For now, this is a placeholder for JWT validation
    // In a production environment, you would validate the JWT here
    
    next();
  } catch (error) {
    logger.error('JWT validation failed', error);
    res.status(401).json({
      message: 'Unauthorized'
    });
  }
};

// Action route handler factory
const createActionHandler = (
  handler: (req: any) => Promise<any>
) => {
  return async (req: express.Request, res: express.Response): Promise<void> => {
    try {
      logger.info('Action request received', {
        action: req.body.action?.name,
        input: JSON.stringify(req.body.input)
      });
      
      const result = await handler(req.body);
      
      res.json(result);
    } catch (error) {
      logger.error('Action handler error', error, {
        action: req.body.action?.name
      });
      
      res.status(400).json({
        message: error instanceof Error ? error.message : 'Unknown error'
      });
    }
  };
};

// Health check endpoint
app.get('/health', (_, res) => {
  res.json({ status: 'ok' });
});

// Root endpoint - documentation
app.get('/', (_, res) => {
  res.json({
    service: 'MamaCare SL Actions Service',
    version: '1.0.0',
    endpoints: [
      '/actions/handlers/calculatePregnancyRisk',
      '/actions/handlers/findNearbyFacilities',
      '/actions/handlers/dispatchAmbulance',
      '/actions/handlers/scheduleAppointment',
      '/actions/handlers/generateMaternalHealthReport'
    ]
  });
});

// Register action handlers
app.post(
  '/actions/handlers/calculatePregnancyRisk',
  validateJwt,
  createActionHandler(calculatePregnancyRisk)
);

app.post(
  '/actions/handlers/findNearbyFacilities',
  createActionHandler(findNearbyFacilities)
);

app.post(
  '/actions/handlers/dispatchAmbulance',
  validateJwt,
  createActionHandler(dispatchAmbulance)
);

app.post(
  '/actions/handlers/scheduleAppointment',
  validateJwt,
  createActionHandler(scheduleAppointment)
);

// Error handling middleware
app.use((
  error: Error,
  _req: express.Request,
  res: express.Response,
  _next: express.NextFunction
) => {
  logger.error('Unhandled error', error);
  
  res.status(500).json({
    message: 'Internal server error'
  });
});

// Start the server
app.listen(PORT, () => {
  logger.info(`MamaCare SL Actions server running on port ${PORT}`);
});

// Handle graceful shutdown
process.on('SIGTERM', () => {
  logger.info('SIGTERM signal received, shutting down');
  process.exit(0);
});

process.on('SIGINT', () => {
  logger.info('SIGINT signal received, shutting down');
  process.exit(0);
});

export default app;
