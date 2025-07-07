import React from 'react';
import { View, Text, StyleSheet, SafeAreaView, ScrollView, TouchableOpacity } from 'react-native';
import { Stack } from 'expo-router';
import { IconSymbol } from '@/components/ui/IconSymbol';
import { Colors } from '@/constants/Colors';
import { useColorScheme } from '@/hooks/useColorScheme';

// Define interfaces for type safety
interface HealthRecord {
  id: string;
  date: string;
  type: 'check-up' | 'test' | 'ultrasound' | 'immunization' | 'medication' | 'measurement';
  title: string;
  description: string;
  provider: string;
  result?: string;
  notes?: string;
}

interface VitalMeasurement {
  id: string;
  date: string;
  type: 'weight' | 'blood-pressure' | 'blood-sugar' | 'heart-rate';
  value: string;
  unit: string;
  status: 'normal' | 'attention' | 'concern';
}

// Mock data for health records
const HEALTH_RECORDS: HealthRecord[] = [
  {
    id: '1',
    date: '2025-05-01',
    type: 'check-up',
    title: 'Monthly Prenatal Visit',
    description: 'Regular prenatal checkup at 24 weeks',
    provider: 'Dr. Sarah Williams',
    notes: 'Everything appears normal. Continue prenatal vitamins.',
  },
  {
    id: '2',
    date: '2025-04-15',
    type: 'test',
    title: 'Blood Work',
    description: 'Complete blood count and screening',
    provider: 'Sierra Leone Medical Laboratory',
    result: 'All levels normal',
    notes: 'Slight iron deficiency. Recommended iron supplement.',
  },
  {
    id: '3',
    date: '2025-04-10',
    type: 'ultrasound',
    title: 'Pregnancy Ultrasound',
    description: '20-week anomaly scan',
    provider: 'Freetown Imaging Center',
    result: 'Normal development',
    notes: 'Baby appears healthy and developing normally. Heart, brain, spine and organs look normal.',
  },
  {
    id: '4',
    date: '2025-03-20',
    type: 'check-up',
    title: 'Monthly Prenatal Visit',
    description: 'Regular prenatal checkup at 16 weeks',
    provider: 'Dr. Mohamed Conteh',
    notes: 'Heartbeat strong. No concerns at this time.',
  },
];

// Mock data for vital measurements
const VITAL_MEASUREMENTS: VitalMeasurement[] = [
  {
    id: '1',
    date: '2025-05-01',
    type: 'weight',
    value: '62.5',
    unit: 'kg',
    status: 'normal',
  },
  {
    id: '2',
    date: '2025-05-01',
    type: 'blood-pressure',
    value: '118/75',
    unit: 'mmHg',
    status: 'normal',
  },
  {
    id: '3',
    date: '2025-04-15',
    type: 'weight',
    value: '61.8',
    unit: 'kg',
    status: 'normal',
  },
  {
    id: '4',
    date: '2025-04-15',
    type: 'blood-pressure',
    value: '125/82',
    unit: 'mmHg',
    status: 'attention',
  },
  {
    id: '5',
    date: '2025-03-20',
    type: 'weight',
    value: '60.2',
    unit: 'kg',
    status: 'normal',
  },
  {
    id: '6',
    date: '2025-03-20',
    type: 'blood-pressure',
    value: '120/78',
    unit: 'mmHg',
    status: 'normal',
  },
];

export default function HealthRecordsScreen(): React.JSX.Element {
  const colorScheme = useColorScheme();
  const [selectedTab, setSelectedTab] = React.useState<'all' | 'tests' | 'vitals'>('all');

  // Function to get icon based on record type
  const getRecordIcon = (type: HealthRecord['type']): string => {
    switch (type) {
      case 'check-up':
        return 'stethoscope';
      case 'test':
        return 'cross.case.fill';
      case 'ultrasound':
        return 'waveform';
      case 'immunization':
        return 'syringe';
      case 'medication':
        return 'pill.fill';
      case 'measurement':
        return 'ruler.fill';
      default:
        return 'doc.text.fill';
    }
  };

  // Function to get color based on measurement status
  const getStatusColor = (status: VitalMeasurement['status']): string => {
    switch (status) {
      case 'normal':
        return '#33B47B';
      case 'attention':
        return '#FFAE33';
      case 'concern':
        return '#EC5766';
      default:
        return '#777';
    }
  };

  // Function to get icon based on measurement type
  const getMeasurementIcon = (type: VitalMeasurement['type']): string => {
    switch (type) {
      case 'weight':
        return 'scalemass.fill';
      case 'blood-pressure':
        return 'heart.fill';
      case 'blood-sugar':
        return 'drop.fill';
      case 'heart-rate':
        return 'waveform.path.ecg';
      default:
        return 'questionmark.circle';
    }
  };

  const renderHealthRecordCard = (record: HealthRecord): React.JSX.Element => {
    const dateObj = new Date(record.date);
    const formattedDate = dateObj.toLocaleDateString('en-US', { 
      month: 'short', 
      day: 'numeric',
      year: 'numeric'
    });

    return (
      <TouchableOpacity
        key={record.id}
        style={styles.recordCard}
        onPress={() => {
          // Navigate to record details in the future
        }}
      >
        <View style={styles.recordHeader}>
          <View style={styles.recordIcon}>
            <IconSymbol size={20} name={getRecordIcon(record.type)} color="#fff" />
          </View>
          <View style={styles.recordHeaderContent}>
            <Text style={styles.recordTitle}>{record.title}</Text>
            <Text style={styles.recordDate}>{formattedDate}</Text>
          </View>
          <IconSymbol size={20} name="chevron.right" color="#999" />
        </View>
        
        <View style={styles.recordDetails}>
          <Text style={styles.recordDescription}>{record.description}</Text>
          <Text style={styles.recordProvider}>{record.provider}</Text>
          
          {record.result && (
            <View style={styles.resultContainer}>
              <Text style={styles.resultLabel}>Result:</Text>
              <Text style={styles.resultText}>{record.result}</Text>
            </View>
          )}
          
          {record.notes && (
            <View style={styles.notesContainer}>
              <Text style={styles.notesLabel}>Notes:</Text>
              <Text style={styles.notesText}>{record.notes}</Text>
            </View>
          )}
        </View>
      </TouchableOpacity>
    );
  };

  const renderVitalMeasurementCard = (): React.JSX.Element => {
    // Group measurements by date
    const measurementsByDate: Record<string, VitalMeasurement[]> = {};
    
    VITAL_MEASUREMENTS.forEach(measurement => {
      if (!measurementsByDate[measurement.date]) {
        measurementsByDate[measurement.date] = [];
      }
      measurementsByDate[measurement.date].push(measurement);
    });

    return (
      <View style={styles.vitalsContainer}>
        {Object.entries(measurementsByDate).map(([date, measurements]) => {
          const dateObj = new Date(date);
          const formattedDate = dateObj.toLocaleDateString('en-US', { 
            month: 'short', 
            day: 'numeric',
            year: 'numeric'
          });

          return (
            <View key={date} style={styles.vitalsByDateCard}>
              <Text style={styles.vitalDateHeader}>{formattedDate}</Text>
              
              <View style={styles.measurementsGrid}>
                {measurements.map(measurement => (
                  <View key={measurement.id} style={styles.measurementItem}>
                    <View style={[
                      styles.measurementIcon,
                      { backgroundColor: getStatusColor(measurement.status) }
                    ]}>
                      <IconSymbol size={16} name={getMeasurementIcon(measurement.type)} color="#fff" />
                    </View>
                    <View style={styles.measurementDetails}>
                      <Text style={styles.measurementType}>
                        {measurement.type.split('-').map(word => 
                          word.charAt(0).toUpperCase() + word.slice(1)
                        ).join(' ')}
                      </Text>
                      <Text style={styles.measurementValue}>
                        {measurement.value} {measurement.unit}
                      </Text>
                    </View>
                  </View>
                ))}
              </View>
            </View>
          );
        })}
      </View>
    );
  };

  return (
    <SafeAreaView style={styles.container}>
      <Stack.Screen options={{ title: 'Health Records' }} />
      
      <View style={styles.header}>
        <Text style={styles.headerTitle}>Health Records</Text>
        <TouchableOpacity style={styles.shareButton}>
          <IconSymbol size={20} name="square.and.arrow.up" color="#555" />
        </TouchableOpacity>
      </View>
      
      <View style={styles.tabContainer}>
        <TouchableOpacity 
          style={[
            styles.tabButton, 
            selectedTab === 'all' && styles.selectedTab
          ]}
          onPress={() => setSelectedTab('all')}
        >
          <Text 
            style={[
              styles.tabText, 
              selectedTab === 'all' && styles.selectedTabText
            ]}
          >
            All Records
          </Text>
        </TouchableOpacity>
        <TouchableOpacity 
          style={[
            styles.tabButton, 
            selectedTab === 'tests' && styles.selectedTab
          ]}
          onPress={() => setSelectedTab('tests')}
        >
          <Text 
            style={[
              styles.tabText, 
              selectedTab === 'tests' && styles.selectedTabText
            ]}
          >
            Tests & Visits
          </Text>
        </TouchableOpacity>
        <TouchableOpacity 
          style={[
            styles.tabButton, 
            selectedTab === 'vitals' && styles.selectedTab
          ]}
          onPress={() => setSelectedTab('vitals')}
        >
          <Text 
            style={[
              styles.tabText, 
              selectedTab === 'vitals' && styles.selectedTabText
            ]}
          >
            Vitals
          </Text>
        </TouchableOpacity>
      </View>
      
      <ScrollView style={styles.scrollView}>
        {(selectedTab === 'all' || selectedTab === 'tests') && (
          <View style={styles.section}>
            {selectedTab === 'all' && <Text style={styles.sectionTitle}>Medical Records</Text>}
            
            <View style={styles.recordsContainer}>
              {HEALTH_RECORDS.map(renderHealthRecordCard)}
            </View>
          </View>
        )}
        
        {(selectedTab === 'all' || selectedTab === 'vitals') && (
          <View style={styles.section}>
            {selectedTab === 'all' && <Text style={styles.sectionTitle}>Vital Measurements</Text>}
            
            {renderVitalMeasurementCard()}
          </View>
        )}
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f8f9fa',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 16,
  },
  headerTitle: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#333',
  },
  shareButton: {
    backgroundColor: '#f0f0f0',
    width: 44,
    height: 44,
    borderRadius: 22,
    justifyContent: 'center',
    alignItems: 'center',
  },
  tabContainer: {
    flexDirection: 'row',
    paddingHorizontal: 16,
    marginBottom: 16,
  },
  tabButton: {
    paddingVertical: 10,
    paddingHorizontal: 12,
    marginRight: 8,
    borderRadius: 20,
  },
  selectedTab: {
    backgroundColor: '#E91E63',
  },
  tabText: {
    fontSize: 14,
    color: '#555',
  },
  selectedTabText: {
    color: '#fff',
    fontWeight: '500',
  },
  scrollView: {
    paddingHorizontal: 16,
  },
  section: {
    marginBottom: 24,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#333',
    marginBottom: 12,
  },
  recordsContainer: {
    marginBottom: 16,
  },
  recordCard: {
    backgroundColor: '#fff',
    borderRadius: 12,
    marginBottom: 16,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 3,
    elevation: 2,
    overflow: 'hidden',
  },
  recordHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#f0f0f0',
  },
  recordIcon: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: '#5E7CE2',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 12,
  },
  recordHeaderContent: {
    flex: 1,
  },
  recordTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#333',
  },
  recordDate: {
    fontSize: 14,
    color: '#777',
    marginTop: 2,
  },
  recordDetails: {
    padding: 16,
  },
  recordDescription: {
    fontSize: 15,
    color: '#444',
    marginBottom: 8,
  },
  recordProvider: {
    fontSize: 14,
    color: '#777',
    marginBottom: 12,
  },
  resultContainer: {
    marginBottom: 8,
  },
  resultLabel: {
    fontSize: 14,
    fontWeight: '600',
    color: '#333',
    marginBottom: 4,
  },
  resultText: {
    fontSize: 14,
    color: '#555',
  },
  notesContainer: {
    marginTop: 4,
  },
  notesLabel: {
    fontSize: 14,
    fontWeight: '600',
    color: '#333',
    marginBottom: 4,
  },
  notesText: {
    fontSize: 14,
    color: '#555',
    lineHeight: 20,
  },
  vitalsContainer: {
    marginBottom: 16,
  },
  vitalsByDateCard: {
    backgroundColor: '#fff',
    borderRadius: 12,
    marginBottom: 16,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 3,
    elevation: 2,
    overflow: 'hidden',
    padding: 16,
  },
  vitalDateHeader: {
    fontSize: 16,
    fontWeight: '600',
    color: '#333',
    marginBottom: 12,
  },
  measurementsGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
  },
  measurementItem: {
    flexDirection: 'row',
    alignItems: 'center',
    width: '50%',
    marginBottom: 12,
  },
  measurementIcon: {
    width: 32,
    height: 32,
    borderRadius: 16,
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 10,
  },
  measurementDetails: {
    flex: 1,
  },
  measurementType: {
    fontSize: 14,
    color: '#555',
  },
  measurementValue: {
    fontSize: 16,
    fontWeight: '600',
    color: '#333',
  },
});
